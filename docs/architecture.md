# Arquitetura do iam

## 1. Visão geral

O `iam` reúne os dois produtos irmãos de identidade, **AuthN** (provedor OIDC)
e **AuthZ** (ponto de decisão AuthZEN), num **monólito modular**: um binário Go
com núcleos hexagonais, implantado como um serviço Cloud Run no laboratório
compartilhado. Os módulos só se falam por interfaces internas e podem virar
serviços separados depois (exige recalcular o teto de conexões).

- Fonte única de verdade: [design aprovado](superpowers/specs/2026-11-06-iam-design.md).
- Decisões: [ADRs](adr/) (índice na seção 3 do design).
- Hospedagem no laboratório: [ADR 0001](adr/0001-lab-hosting.md).

Este documento ilustra o design; não acrescenta decisões. Onde algo é do marco 2
ou de camada posterior, a legenda diz.

## 2. C4 nível 1: contexto

```mermaid
flowchart TB
  user["Usuário final<br/>(navegador)"]
  op["Master e operadores<br/>(tenant de sistema)"]
  a["Cliente de teste A<br/>web BFF em Go (examples/a)"]
  subgraph iam["iam (um implantável)"]
    authn["AuthN<br/>provedor OIDC + API de login guiada"]
    authz["AuthZ<br/>PDP AuthZEN + relações e políticas"]
  end
  idp["IdP externo<br/>(qualquer emissor OIDC)"]
  smtp["Provedor SMTP"]
  db[("Cloud SQL lab-postgres<br/>compartilhado, dono: aleogr/lab")]
  gha["GitHub Actions<br/>CI e Terraform"]

  user -->|"login OIDC"| authn
  user -->|"usa"| a
  a -->|"OIDC code + PKCE (C8)"| authn
  a -->|"AuthZEN via SDK Go (C7)"| authz
  op -->|"API administrativa (C6)"| iam
  authz -.->|"só pela interface pública: token OIDC ou sujeito AuthZEN"| authn
  idp -.->|"tokens aceitos como sujeito"| authz
  authn -->|"porta shared/messaging, adaptador SMTP"| smtp
  iam -->|"banco iam"| db
  gha -->|"Workload Identity Federation"| iam
```

_Legenda: AuthN e AuthZ formam um único implantável; o AuthZ também funciona com um IdP externo, sem login pelo AuthN._

## 3. C4 nível 2: containers no laboratório

```mermaid
flowchart LR
  subgraph gh["GitHub"]
    ci["GitHub Actions<br/>iam-planner (PRs, leitura)<br/>iam-deployer (só main, escrita)"]
  end
  subgraph proj["Projeto GCP aleogr-lab-iam-c4vc"]
    svc["Cloud Run service iam<br/>conta iam-run<br/>pool máx. 2, 1 instância"]
    job["Cloud Run job de migração<br/>conta iam-migrate<br/>1 conexão"]
    sm["Secret Manager<br/>senha do iam_migrator"]
  end
  subgraph shared["Projeto aleogr-lab-shared-dacd (aleogr/lab)"]
    subgraph pg["lab-postgres (Postgres 16, db-f1-micro)"]
      dbiam[("banco iam<br/>CONNECTION LIMIT 6")]
    end
  end
  dns["Cloudflare DNS<br/>(futuro, sob lab.aleogr.dev)"]

  ci -->|"WIF, terraform plan/apply"| proj
  svc -->|"usuário IAM iam-run<br/>membro de iam_app (sem senha)"| dbiam
  job -->|"iam_migrator<br/>dono do schema"| dbiam
  sm -->|"senha"| job
  dns -.->|"futuro"| svc
```

_Legenda: o aleogr/lab é dono da instância; o iam declara apenas o banco `iam` e os usuários `iam`/`iam_*`. A instância dorme seg–qui, 22:00–07:30 (São Paulo)._

| Identidade no banco | Tipo | Permissões |
|---|---|---|
| `iam_migrator` | Usuário com senha (Secret Manager; criado à mão, importado pelo Terraform) | Dono do schema; aplica migrações |
| `iam_app` | Role sem login | Só DML; sujeita a RLS; sem `BYPASSRLS`; não é dona das tabelas |
| `iam-run` | Usuário IAM, sem senha | Membro de `iam_app`; usado pelo serviço |

Bootstrap SQL: `REVOKE CONNECT ... FROM PUBLIC`, `GRANT CONNECT` por nome,
`CONNECTION LIMIT 6`. Contas de runtime: `iam-release`, `iam-run`, `iam-migrate`.

## 4. C4 nível 3: componentes do binário (marco 1)

```mermaid
flowchart TB
  cmd["cmd/iam<br/>config, wiring, servidores HTTP"]
  subgraph servers["Servidores (bordas)"]
    authns["authn/server<br/>OIDC + API de login guiada"]
    loginui["authn/login-ui"]
    authzs["authz/server<br/>AuthZEN, relações, políticas"]
    api["api/ (Protobuf, Connect)"]
  end
  subgraph cores["Núcleos hexagonais"]
    authnc["authn/core<br/>contas, credenciais, sessão, fluxo de login"]
    authzc["authz/core<br/>relações, políticas, avaliador, explicação"]
  end
  subgraph sharedlib["shared/"]
    cfg["config"]
    ten["tenancy"]
    aud["audit"]
    sm["statemachine"]
    subgraph ports["Portas"]
      st["storage"]
      cr["crypto<br/>(única entrada de criptografia)"]
      msg["messaging"]
    end
  end
  subgraph adapters["Adaptadores"]
    pga["adapters/postgres<br/>RLS, migrações"]
    smtpa["adaptador SMTP"]
  end
  sdk["sdks/go<br/>cliente AuthZEN com cache, validação de token"]
  ex["examples/a"]

  cmd --> servers
  cmd --> adapters
  cmd --> cfg
  authns --> authnc
  loginui -->|"API guiada (C9)"| authns
  authzs --> authzc
  api --> authns
  api --> authzs
  authnc --> ports
  authzc --> ports
  authnc --> sm
  authnc --> ten
  authnc --> aud
  authzc --> ten
  authzc --> aud
  pga -.->|"implementa"| st
  smtpa -.->|"implementa"| msg
  ex --> sdk
  sdk -->|"HTTP (C7, C8)"| servers
```

_Legenda: setas cheias são dependências permitidas; tracejadas, implementação de porta. Nenhuma seta sai de um núcleo para servidor ou adaptador._

| Regra de dependência | Consequência |
|---|---|
| Núcleos dependem só de portas e de bibliotecas puras de `shared/` | `authn/core` e `authz/core` não conhecem HTTP, banco nem nuvem |
| Adaptadores implementam portas | `adapters/postgres` implementa `shared/storage`; SMTP implementa `shared/messaging` |
| `shared/crypto` é a única entrada de criptografia | Hash de senha, JOSE (`go-jose`), aleatoriedade, comparação em tempo constante |
| AuthZ consome o AuthN só pela interface pública | Token OIDC ou sujeito AuthZEN; nada de importar `authn/core` |
| `examples/` só usa APIs públicas | Módulo separado, sem acesso a `internal` |

Módulos publicáveis com `go.mod` próprio: `authn/core`, `authz/core`,
`shared/statemachine`, SDKs Go; `go.work` na raiz. A divisão exata é decisão de
implementação no marco 0.

## 5. Fronteiras verificadas no CI

| Regra | Verificação |
|---|---|
| Núcleos não importam servidores, adaptadores, HTTP, SQL nem SDK de nuvem | `depguard` |
| Primitivas criptográficas, `go-jose` e `crypto/*` só em `shared/crypto` | `depguard` + `forbidigo` |
| Aleatoriedade só via `crypto/rand`; relógio e aleatoriedade injetados nos núcleos | `forbidigo` (`time.Now`, `math/rand` fora de adaptadores) |
| Projetos de teste (`examples/`) só usam APIs públicas | Módulo separado sem acesso a pacotes `internal` |
| Toda consulta passa pelo escopo de tenant | Porta de armazenamento exige `TenantID`; testes de abuso entre tenants |

## 6. Hierarquia de tenants

```mermaid
flowchart TB
  inst["Instalação"]
  sys["Tenant de sistema<br/>invisível aos tenants"]
  master["Master<br/>token de uso único + MFA resistente a phishing"]
  ops["Operadores<br/>operação do dia a dia"]
  t["Tenant<br/>issuer https://host/t/tenant"]
  org1["Organização"]
  org2["Organização"]
  u["Usuário<br/>uma pessoa, uma conta"]
  m1["Membership"]
  m2["Membership"]
  app["Aplicação<br/>cliente OAuth, público ou confidencial, PKCE sempre"]

  inst --> sys
  sys --> master
  sys --> ops
  inst --> t
  t --> org1
  t --> org2
  t --> u
  t --> app
  u --> m1 --> org1
  u --> m2 --> org2
```

_Legenda: o usuário pertence ao tenant e entra em várias organizações por memberships; papéis por organização são relações no AuthZ._

```mermaid
flowchart LR
  i["Instância"] --> te["Tenant"] --> o["Organização"] --> ap["Aplicação"]
  floor["Pisos de segurança<br/>+ proibições absolutas (ADR 0012)"] -.->|"níveis inferiores não rebaixam"| ap
```

_Legenda: herança de configuração; cada nível pode endurecer, nunca rebaixar os pisos; alertas exigem ciência registrada e há modo estrito._

## 7. Fluxos

### 7.1 Login OIDC (authorization code + PKCE) com senha ou passkey

```mermaid
sequenceDiagram
  autonumber
  actor U as Usuário (navegador)
  participant A as Cliente A (BFF)
  participant UI as authn/login-ui
  participant S as authn/server, issuer host/t/tenant
  participant C as authn/core
  participant K as shared/crypto
  U->>A: acessa área protegida
  A->>U: redirect para /t/tenant/authorize com PKCE S256, state, nonce
  U->>S: GET authorize
  S->>C: cria fluxo de login (máquina de estados)
  S->>UI: exibe login
  UI->>S: API guiada: estado e próximos passos
  S-->>UI: passo identificação
  UI->>S: identificador
  S->>C: transição identificação para fator primário
  S-->>UI: passos: senha ou passkey
  alt senha
    UI->>S: senha
    C->>K: verifica Argon2id
  else passkey
    S-->>UI: desafio WebAuthn (RP ID = host do AuthN)
    UI->>S: asserção
    C->>K: verifica assinatura, credencial do mesmo tenant
  end
  C->>C: sessão registra fator e horário
  opt política exige segundo fator
    S-->>UI: passo segundo fator
    UI->>S: segundo fator
    C->>C: sessão registra fator e horário
  end
  C->>C: transição para autenticado (cada transição gera evento de auditoria)
  S->>U: cookie __Host- HttpOnly Secure e redirect com code e state
  U->>A: callback com code
  A->>S: POST token com code_verifier
  S-->>A: access token JWS ES256, ID token, refresh token opaco
  A->>A: valida ID token via JWKS (SDK Go)
```

_Legenda: marco 1. O servidor valida cada transição; não há como chegar a "autenticado" sem os fatores exigidos._

### 7.2 Decisão AuthZEN do cliente A

```mermaid
sequenceDiagram
  autonumber
  participant A as Cliente A
  participant SDK as SDK Go (cache)
  participant P as authz/server (PDP)
  participant Z as authz/core
  A->>SDK: evaluation(sujeito, ação, recurso, contexto)
  alt decisão em cache dentro do TTL
    SDK-->>A: decisão em cache
  else sem cache válido
    SDK->>P: AuthZEN evaluation
    alt PDP responde
      P->>Z: avalia políticas e relações do tenant
      alt erro no meio da avaliação
        Z-->>P: nega (fail-closed)
      else
        Z-->>P: permit ou deny + auditoria da decisão
      end
      P-->>SDK: decisão
      SDK->>SDK: guarda no cache com TTL curto configurável
      SDK-->>A: decisão
    else PDP fora do ar e sem cache dentro do TTL
      SDK-->>A: nega (fail-closed)
    end
  end
```

_Legenda: marco 1 (avaliação individual e em lote, cache de decisões). Avaliação local no SDK fica para o marco 2._

### 7.3 Escrita de relação com outbox e token de consistência

```mermaid
sequenceDiagram
  autonumber
  participant App as Aplicação cliente
  participant DB as Banco da aplicação
  participant R as Relay do SDK (polling)
  participant P as authz/server
  App->>DB: transação: dado de negócio + linha no outbox (SDK)
  DB-->>App: commit
  loop polling
    R->>DB: lê pendentes do outbox
    R->>P: escreve relação (idempotente)
    P-->>R: token de consistência
    R->>DB: marca enviado e guarda token
  end
  App->>P: check com token (pelo menos tão atual quanto)
  P-->>App: decisão que já enxerga a escrita
```

_Legenda: marco 2 (ADR 0016). Tokens de consistência servem às revogações sensíveis; no resto, consistência eventual configurável._

### 7.4 Aprovação M de N

```mermaid
sequenceDiagram
  autonumber
  actor Sol as Solicitante
  participant Z as AuthZ (solicitações)
  actor Ap as Aprovador (cada um de M)
  participant E as Executor
  Sol->>Z: solicita ação
  Z->>Z: ação canônica (RFC 8785) e hash
  loop até o quórum, dentro da janela
    Z-->>Ap: mostra a mesma representação canônica
    Z-->>Ap: desafio WebAuthn = hash + nonce
    Ap->>Z: assinatura com verificação de usuário
    Z->>Z: valida pessoa distinta, não é o solicitante, SoD, papéis do quórum
  end
  Z->>Z: assinaturas guardadas na auditoria
  alt execução por reavaliação
    E->>Z: AuthZEN check com a aprovação
    Z-->>E: permit
  else pacote de evidências
    Z-->>E: ação canônica + assinaturas
    E->>E: verifica as assinaturas por conta própria
  end
```

_Legenda: camada posterior, fora do marco 1 (ADR 0021)._

## 8. Máquinas de estados

Na implementação estes diagramas serão gerados a partir das definições das
máquinas de estados no código (brief 4.4); aqui está o design pretendido.
Transições não previstas são recusadas; cada transição usa versão otimista e
gera evento de auditoria.

```mermaid
stateDiagram-v2
  state "Identificação" as Ident
  state "Fator primário" as Prim
  state "Segundo fator" as Seg
  state "Autenticado" as Auth
  [*] --> Ident
  Ident --> Prim: identificador aceito
  Prim --> Seg: fator cumprido e política exige outro
  Prim --> Auth: fator cumprido e política satisfeita
  Seg --> Auth: fator cumprido
  Auth --> [*]
```

_Legenda: fluxo de login (marco 1); não existe transição que pule etapas._

```mermaid
stateDiagram-v2
  state "Solicitada" as Sol
  state "Aprovada" as Apr
  state "Negada" as Neg
  state "Ativa" as Atv
  state "Expirada" as Exp
  state "Revogada" as Rev
  [*] --> Sol
  Sol --> Apr
  Sol --> Neg
  Apr --> Atv: gera relação temporária
  Atv --> Exp: fim da validade
  Atv --> Rev
  Neg --> [*]
  Exp --> [*]
  Rev --> [*]
```

_Legenda: ciclo de vida da solicitação de acesso; camada posterior (ADR 0020)._

## 9. Dados

PostgreSQL 16 (`lab-postgres`), driver `pgx`. Sem Redis: sessões, rate limiting
e filas no Postgres. Migrações nomeadas por data e hora, sempre expand/contract,
aplicadas pelo job como `iam_migrator`.

| Módulo | Tabelas principais (nomes indicativos) | Escopo |
|---|---|---|
| `shared/tenancy` | tenants, organizações, usuários, memberships, aplicações | `tenant_id` |
| `shared/config` | configuração por nível (instância, tenant, organização, aplicação), cientes de alertas | `tenant_id` |
| `shared/audit` | eventos de auditoria encadeados por hash, só inclusão | `tenant_id` |
| `shared/statemachine` | instâncias de fluxo com estado e versão | `tenant_id` |
| `authn` | credenciais (senha, passkey), sessões com fatores e horários, códigos de autorização, refresh tokens, rate limiting | `tenant_id` |
| `authz` | relações, versões de políticas, decisões auditadas | `tenant_id` |

- Isolamento na camada de aplicação (porta exige `TenantID`), com **RLS como segunda barreira**.
- `iam_app` não é dono das tabelas e não tem `BYPASSRLS`.
- Nada de `LISTEN/NOTIFY` nem conexão dedicada; filas e outbox por polling;
  pool esgotado espera com timeout, nunca abre conexão extra.

| Consumidor | Conexões |
|---|---|
| Serviço: pool 2 × 1 instância × 2 revisões no rollout | 4 |
| Job de migração | 1 |
| Uma pessoa no psql | 1 |
| Tarefas agendadas (usam o pool do serviço via endpoint) | 0 |
| **Total (teto, com teste que falha acima disso)** | **6** |

## 10. Contratos

Versionados, cobertos por testes de contrato, alterados só por mudança
explícita e revisada.

| Contrato | Forma | Libera em paralelo |
|---|---|---|
| C1. Esquema de configuração | Tipos Go + validação + documentação gerada | Todos os módulos que leem configuração |
| C2. Porta de criptografia | Interface Go em `shared/crypto` | Senhas, tokens, chaves, passkeys |
| C3. Portas de armazenamento + suíte de contrato | Interfaces Go por domínio + testes de contrato | Adaptador Postgres × núcleos; abuso entre tenants |
| C4. Evento de auditoria | Esquema versionado | Auditoria × todos os emissores |
| C5. Biblioteca de máquinas de estados | API Go | Fluxo de login, sessão, solicitações |
| C6. API administrativa | Protobuf (`api/`), `buf breaking` | Servidor × console × SDKs gerados |
| C7. AuthZEN | Especificação 1.0 + esquema JSON | `authz/server` × SDK Go × cliente A |
| C8. OIDC do AuthN | Especificação OIDC + discovery | `authn/server` × cliente A × SDK |
| C9. API de login guiada | Esquema de estado e passos | `authn/server` × `login-ui` × futuros apps nativos |
| C10. Linguagem de políticas e modelo de relações | Gramática + esquema JSON + corpus de testes | Avaliador × oráculo × SDK |

```mermaid
flowchart LR
  C1["C1 config"] --> W1["módulos que leem configuração"]
  C2["C2 crypto"] --> W2["senhas, tokens, chaves, passkeys"]
  C3["C3 storage"] --> W3["adapters/postgres, núcleos, abuso entre tenants"]
  C4["C4 auditoria"] --> W4["shared/audit, emissores"]
  C5["C5 statemachine"] --> W5["login, sessão, solicitações"]
  C6["C6 API admin"] --> W6["servidor, console, SDKs gerados"]
  C7["C7 AuthZEN"] --> AS["authz/server"]
  C8["C8 OIDC"] --> NS["authn/server"]
  C9["C9 login guiado"] --> NS
  C9 --> UI["login-ui, futuros apps nativos"]
  C10["C10 políticas e relações"] --> EV["avaliador, oráculo cedar-go"]
  C7 & C8 & C10 --> SDK["SDK Go"]
  C7 & C8 --> CA["cliente A"]
```

_Legenda: cada contrato fixado libera, em paralelo, as frentes que ficam dos dois lados dele._
