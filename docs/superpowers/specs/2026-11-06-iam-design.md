# Design do iam (AuthN e AuthZ): fundação, marco 0 e marco 1

- Data: 2026-11-06
- Status: em revisão pelo dono
- Origem: brainstorming de 2026-11-06 sobre o `docs/brief.md`, com base no
  `docs/market-study/`. Cada decisão aprovada tem um ADR em `docs/adr/`.

Este documento consolida o design. O brief continua sendo o material de
partida e o destino "completo"; aqui fica o que foi decidido, o que muda em
relação ao brief e o que os dois primeiros marcos entregam.

## 1. Objetivo

Construir os dois produtos irmãos de identidade (AuthN e AuthZ) como o
primeiro projeto de infraestrutura de uma família de projetos de
infraestrutura, cada um no seu repositório, consumidos por projetos finais que
ainda não existem. O destino é o escopo completo do brief (seção 5), alcançado
em camadas prontas, testadas e auditadas.

**Sucesso nesta fase:** os marcos 0 e 1 concluídos com os critérios da seção 12,
e o roadmap das camadas seguintes definido com os contratos que liberam frentes
em paralelo.

## 2. Contexto e restrições

| Restrição | Origem |
|---|---|
| Todo o código é escrito por agentes; o dono define, opera e aprova | Premissa 14 |
| Dono revisa mais de 10 h por semana; o gargalo é a verificação automatizada | ADR 0004 |
| Sem consumidor real: A, B, C são clientes de teste fictícios | ADR 0002 |
| Custo mínimo: laboratório na instância compartilhada `lab-postgres` (6 conexões, dorme seg–qui 22:00–07:30) | ADR 0001 |
| Repositório público, Apache-2.0, nenhum segredo versionado | ADR 0009 |
| Premissas 1 a 26 do brief valem para os dois produtos | Brief, seção 2 |

## 3. Decisões

| ADR | Decisão |
|---|---|
| [0001](../../adr/0001-lab-hosting.md) | O iam no laboratório compartilhado (projeto, federação, banco, teto de 6 conexões) |
| [0002](../../adr/0002-test-client-projects.md) | Projetos A, B, C são clientes de teste fictícios; nenhum consumidor real por enquanto |
| [0003](../../adr/0003-one-repository-per-infrastructure-product.md) | Um repositório por produto de infraestrutura |
| [0004](../../adr/0004-milestones-0-and-1.md) | Marco 0 "fábrica" e marco 1 AuthN e AuthZ mínimos ponta a ponta |
| [0005](../../adr/0005-state-tables-and-hash-chained-audit.md) | Estado em tabelas e log de auditoria encadeado; sem event sourcing |
| [0006](../../adr/0006-admin-api-protobuf-connect.md) | API administrativa em Protobuf, servida com Connect, OpenAPI gerada |
| [0007](../../adr/0007-path-based-tenant-issuer.md) | Issuer por tenant no caminho da URL |
| [0008](../../adr/0008-single-product-version.md) | Versão única dos componentes de produto |
| [0009](../../adr/0009-apache-2-license.md) | Licença Apache-2.0 |
| [0010](../../adr/0010-jose-via-go-jose.md) | Exceção de segurança da premissa 24 cobre JOSE: `go-jose` v4 |
| [0011](../../adr/0011-latency-and-scale-targets.md) | Metas de latência e de escala |
| [0012](../../adr/0012-absolute-prohibitions.md) | Proibições absolutas e alertas críticos |
| [0013](../../adr/0013-authz-hybrid-model.md) | AuthZ: modelo híbrido (relações + políticas) |
| [0014](../../adr/0014-authz-own-engine-with-oracle.md) | AuthZ: motor próprio, `cedar-go` como oráculo de teste |
| [0015](../../adr/0015-authz-cedar-core-language.md) | AuthZ: linguagem com núcleo compatível com Cedar |
| [0016](../../adr/0016-authz-relation-sync-and-consistency.md) | AuthZ: escrita de relações, outbox e tokens de consistência |
| [0017](../../adr/0017-authz-local-evaluation-and-cache.md) | AuthZ: PDP central, cache e avaliação local no SDK |
| [0018](../../adr/0018-authz-listing.md) | AuthZ: plano de filtro e listagem paginada |
| [0019](../../adr/0019-authz-policy-lifecycle.md) | AuthZ: explicabilidade, simulação, testes, versões e governança |
| [0020](../../adr/0020-authz-dynamic-access.md) | AuthZ: acesso temporário, solicitações, delegação e obrigações |
| [0021](../../adr/0021-authz-multi-party-approval.md) | AuthZ: aprovação M de N |
| [0022](../../adr/0022-authz-states-and-transitions.md) | AuthZ: estados e transições declarativos |
| [0023](../../adr/0023-authz-machines-and-agents.md) | AuthZ: máquinas e agentes |
| [0024](../../adr/0024-authz-standards.md) | AuthZ: padrões adotados |
| [0025](../../adr/0025-authz-milestone-1-scope.md) | AuthZ: escopo no marco 1 |
| [0026](../../adr/0026-no-own-authenticator-app.md) | Sem app autenticador próprio |
| [0027](../../adr/0027-session-factors-and-nist-passwords.md) | Sessão com fatores carimbados e regras de senha do NIST SP 800-63-4 |
| [0028](../../adr/0028-cyber-resilience-act-from-milestone-0.md) | Obrigações do Cyber Resilience Act desde o marco 0 |
| [0029](../../adr/0029-macos-not-blocked.md) | macOS não é bloqueado pelo design |

## 4. Arquitetura

### 4.1 Forma

- **Monólito modular:** um binário (`iam`) com AuthN e AuthZ, implantado no
  Cloud Run; os módulos só se falam por interfaces internas e podem virar
  serviços separados depois (exige recalcular o teto de conexões, ADR 0001).
- **Núcleos hexagonais:** `authn/core` e `authz/core` não conhecem HTTP, banco
  nem nuvem; dependem de portas.
- **Dois produtos, um modelo de tenants:** tenant → organização → membership,
  compartilhado (`shared/tenancy`).
- **AuthZ consome o AuthN só pela interface pública** (token OIDC ou sujeito
  AuthZEN), e funciona com IdP externo (premissa 8).

### 4.2 Estrutura do repositório

Igual à seção 4.1 do brief, com estas mudanças:

| Mudança | Motivo |
|---|---|
| Sem `notifications/` | ADR 0003; fica a porta `shared/messaging` com adaptador SMTP direto |
| `cmd/iam/` | O binário do monólito modular |
| `shared/statemachine/` publicado como módulo próprio | Reuso por outros produtos de infraestrutura (ADR 0003, 0022) |
| `deploy/gcp/` | Infraestrutura do laboratório (ADR 0001), já existente |

Um `go.mod` por componente publicável (`authn/core`, `authz/core`,
`shared/statemachine`, SDKs Go); o resto vive no módulo do servidor. `go.work`
na raiz para desenvolvimento. A divisão exata é decisão de implementação no
marco 0, guiada pela regra: quem importa uma biblioteca não herda dependências
de servidor.

### 4.3 Fronteiras verificadas no CI

| Regra | Verificação |
|---|---|
| Núcleos não importam servidores, adaptadores, HTTP, SQL nem SDK de nuvem | `depguard` |
| Primitivas criptográficas, `go-jose` e `crypto/*` só em `shared/crypto` | `depguard` + `forbidigo` |
| Aleatoriedade só via `crypto/rand`; relógio e aleatoriedade injetados nos núcleos | `forbidigo` (`time.Now`, `math/rand` fora de adaptadores) |
| Projetos de teste (`examples/`) só usam APIs públicas | Módulo separado sem acesso a pacotes `internal` |
| Toda consulta passa pelo escopo de tenant | Porta de armazenamento exige `TenantID`; testes de abuso entre tenants |

### 4.4 Componentes do marco 1

```
cmd/iam                 binário: config, wiring, servidores HTTP
shared/config           esquema tipado, validação, precedência, alertas e proibições (ADR 0012)
shared/crypto           única porta de criptografia: hash de senha, JOSE, aleatoriedade, comparação em tempo constante
shared/tenancy          tenant, organização, membership, aplicação (cliente OAuth)
shared/audit            eventos versionados, encadeamento por hash, somente inclusão
shared/statemachine     tabela de transições com guards, versão otimista, evento por transição
shared/storage          portas de armazenamento orientadas ao domínio + testes de contrato
adapters/postgres       implementação das portas, RLS como segunda barreira, migrações
authn/core              contas, credenciais (senha, passkey), sessão com fatores carimbados, fluxo de login
authn/server            OIDC: discovery, authorize (PKCE), token, JWKS, userinfo; API de login guiada pelo servidor
authn/login-ui          telas mínimas de login e cadastro sobre a API guiada
authz/core              relações, políticas (núcleo Cedar), avaliador, explicação
authz/server            AuthZEN (avaliação individual e em lote), API de relações e políticas
api/                    Protobuf da API administrativa
sdks/go                 cliente AuthZEN com cache; validação de token
examples/a              cliente de teste A (web BFF em Go)
e2e/                    cenário ponta a ponta do marco 1
```

## 5. Modelo de domínio compartilhado

| Conceito | Regra |
|---|---|
| Tenant | Isolamento total. Todo dado e toda consulta carregam `tenant_id`. Issuer `https://<host>/t/<tenant>` (ADR 0007) |
| Organização | Dentro de um tenant; políticas, administradores e branding próprios |
| Usuário | Pertence ao tenant; uma pessoa, uma conta, várias organizações |
| Membership | Liga usuário a organização; papéis por organização são relações no AuthZ |
| Aplicação | Cliente OAuth do tenant; tipo público ou confidencial; PKCE sempre |
| Sessão central | Registra cada fator cumprido com horário (ADR 0027); política exige fatores e frescor |
| Tenant de sistema | Hospeda o master e os operadores; invisível aos tenants (4.23) |

Configuração com herança instância → tenant → organização → aplicação, pisos de
segurança que níveis inferiores não rebaixam, alertas com ciência registrada e
modo estrito (brief 4.5 e 4.6), e as proibições absolutas do ADR 0012.

## 6. AuthN no marco 1

| Item | Decisão |
|---|---|
| Protocolo | Provedor OIDC com authorization code + PKCE (S256), discovery, JWKS, userinfo. Sem implicit e ROPC (ADR 0012) |
| Login | API guiada pelo servidor: devolve o estado do fluxo e os próximos passos (4.34); cada fluxo é uma máquina de estados validada no servidor |
| Métodos | Senha (Argon2id, regras do NIST SP 800-63-4, checagem de senha vazada por lista local) e passkey (WebAuthn, `go-webauthn`), RP ID = host do AuthN |
| Isolamento de passkey | Credencial registrada num tenant é recusada em qualquer outro (teste de abuso obrigatório, ADR 0007) |
| Tokens | Access token JWS (ES256), ID token, refresh token opaco com rotação e detecção de reuso; chaves em memória carregadas do adaptador de chaves |
| Sessão | Cookie `__Host-`, `HttpOnly`, `Secure`; proteção anti-CSRF por token e `Origin`; `Cache-Control: no-store` |
| Master | Inicialização por token de uso único; MFA resistente a phishing obrigatório (ADR 0012) |
| Fora do marco 1 | Federação, SAML, LDAP, ICP-Brasil, TOTP, links mágicos, SMS/WhatsApp, device flow, CIBA, DPoP, SCIM, consentimento, app nativo |

## 7. AuthZ

### 7.1 Modelo e motor (ADRs 0013 a 0015)

- **Híbrido:** o AuthZ guarda relações (`organização#membro@usuário`,
  `paciente#atendido_por@usuário`), e políticas `permit`/`forbid` usam
  relações, atributos e contexto. RBAC é caso particular.
- **Motor próprio:** armazém de relações no Postgres (porta), avaliador e
  listagem próprios, sem dependência no runtime.
- **Linguagem:** núcleo compatível com Cedar (sem laços, regex nem ponto
  flutuante); extensões explícitas para validade de relações e transições,
  traduzíveis para o núcleo nos testes. Representação JSON com a mesma
  validação do texto.
- **Verificação:** testes diferenciais contra `cedar-go` (oráculo, só em
  teste), fuzzing de parser e avaliador com limites de tamanho e profundidade,
  testes de propriedade (check, lote e listagem coerentes), fail-closed em
  qualquer erro no meio da avaliação, auditoria externa antes de produção.

### 7.2 Dados, consistência, desempenho (ADRs 0016 a 0018)

- Escrita de relações pela API, com outbox transacional no SDK e relay
  idempotente; tokens de consistência para revogações sensíveis.
- PDP central; o SDK avalia localmente o que não depende de relações e guarda
  em cache decisões relacionais com TTL curto configurável; PDP fora do ar →
  cache dentro do TTL, depois nega.
- Listagem por plano de filtro (avaliação parcial → predicado que o SDK traduz
  para SQL) e listagem paginada no AuthZ para o que depende só das relações
  dele; nunca trunca; listagem nunca mais permissiva que o check.

### 7.3 Ciclo de vida e acesso dinâmico (ADRs 0019 a 0023)

- Explicação de decisão (detalhe só para administração), simulação "e se?",
  testes de políticas por CLI, versões imutáveis, modo sombra antes de ativar,
  pisos por hierarquia (`forbid` superior vence).
- Relações com validade; solicitações de acesso como máquina de estados;
  delegação só de subconjunto, com prazo; obrigações (aplicação que não entende
  nega); obrigação de step-up liga ao `acr` do AuthN.
- M de N com quórum por papéis e assinatura WebAuthn sobre a ação canônica
  (RFC 8785), execução por reavaliação ou pacote de evidências.
- Estados e transições declarativos com definição única para AuthZ e
  aplicação; ações permitidas na resposta.
- Agentes e máquinas como sujeitos; atuação em nome de alguém por delegação e
  token exchange (RFC 8693).

### 7.4 Padrões (ADR 0024)

AuthZEN 1.0 (avaliação, lote, busca); perfil de solicitação e aprovação do
AuthZEN; RAR nos tokens do AuthN; escopos OAuth para permissões amplas. UMA 2.0
fora; XACML só como referência.

### 7.5 AuthZ no marco 1 (ADR 0025)

Armazém de relações, políticas no núcleo Cedar com versões validadas, papéis
por organização, AuthZEN avaliação individual e em lote, explicação básica,
auditoria de decisões, SDK Go com cache de decisões. A plataforma usa esse
RBAC nos papéis administrativos (operador, administrador de tenant e de
organização). Plano de filtro, avaliação local, outbox e tokens de consistência
vão para o marco 2.

## 8. Contratos

Contratos são versionados, cobertos por testes de contrato e mudam só por
mudança explícita e revisada. Cada um libera frentes de trabalho:

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

## 9. Dados e persistência

- PostgreSQL 16 (o do `lab-postgres`); driver `pgx` (dependência justificada:
  implementar o protocolo do Postgres foge do escopo).
- `tenant_id` em toda tabela; RLS como segunda barreira, com o papel `iam_app`
  sem `BYPASSRLS` e sem posse das tabelas (ADR 0001).
- Migrações próprias, nomeadas por data e hora, sempre expand/contract,
  aplicadas pelo job de migração como `iam_migrator`.
- Sem Redis: sessões, rate limiting e filas no Postgres.
- **Orçamento de conexões:** pool máximo 2 no serviço, 1 instância, 1 no job de
  migração; nada de `LISTEN/NOTIFY` nem conexão dedicada; constante de
  orçamento com teste que falha acima de 6 (ADR 0001).

## 10. Segurança

- Proibições absolutas e alertas críticos do ADR 0012, verificados ao carregar
  a configuração e por API.
- Camada criptográfica única (`shared/crypto`), regra no CI proibindo
  primitivas fora dela.
- **Política de dependências** (premissa 24 e brief 4.30). Aceitas no runtime,
  com justificativa:

| Dependência | Motivo |
|---|---|
| Biblioteca padrão e `golang.org/x/{crypto,text,net}` | Base |
| `github.com/go-jose/go-jose/v4` (≥ 4.1.8) | JOSE (ADR 0010) |
| `github.com/go-webauthn/webauthn` | Formatos binários do WebAuthn (exceção de segurança do brief 4.30); v0.x fixada e revisada a cada atualização |
| `github.com/jackc/pgx/v5` | Protocolo do Postgres |
| `google.golang.org/protobuf`, `connectrpc.com/connect` | API administrativa (ADR 0006) |

  Só em teste: `github.com/cedar-policy/cedar-go` (oráculo). Qualquer outra
  dependência exige justificativa registrada no PR.

- Modelo de ameaças inicial em `docs/threat-model.md`; cada ameaça vira teste
  de abuso.

## 11. Verificação

| Camada | O que roda |
|---|---|
| Cada PR | `gofmt`, `golangci-lint` (complexidade, duplicação, `depguard`, `forbidigo`, `gosec`), `go vet`, testes com `-race`, testes de contrato contra Postgres real, fuzzing curto, `govulncheck`, `gitleaks`, verificação de licenças, `buf lint` e `buf breaking`, código gerado atualizado, `actionlint`, varredura da imagem |
| Noite | Fuzzing longo, testes diferenciais amplos contra o oráculo, benchmarks com detecção de regressão (ADR 0011) |
| Antes de release | Suíte de conformidade OIDC e AuthZEN, e2e completo, SBOM, assinatura e proveniência |
| Revisão | Agente revisor que não escreveu o código, com checklist de qualidade (brief 4.30) e de OWASP; agente "atacante" nas áreas sensíveis |

Testes de abuso obrigatórios no marco 1: `alg: none` e confusão de algoritmo;
token de outro tenant; `redirect_uri` manipulada; ausência de `state`, `nonce`
ou PKCE; passkey e sessão de outro tenant; enumeração de usuários por mensagem
ou tempo; força bruta; reutilização de código de autorização e de refresh token
(concorrência); fixação de sessão; CSRF; escalada de administrador de tenant
para operador; decisão de AuthZ entre tenants; erro no meio da avaliação nega.

## 12. Marcos

### Marco 0: fábrica

**Objetivo:** tudo o que torna a verificação automatizada confiável, antes de
código de produto.

| Entrega | Critério de conclusão |
|---|---|
| Estrutura de módulos e `go.work` | Compila; fronteiras verificadas por `depguard` com um teste que prova que uma violação falha |
| CI geral | Todos os checks da seção 11 (camada "cada PR") rodando e obrigatórios no ruleset |
| Ferramentas de contrato | `buf` configurado; um Protobuf mínimo passa por lint, breaking e geração |
| Ambiente local | Um comando sobe Postgres e o servidor vazio com health check |
| Harness e2e | O cliente de teste A sobe e conversa com o servidor vazio no CI |
| Comunidade e CRA | `LICENSE`, `SECURITY.md` (com o canal privado do GitHub), `CONTRIBUTING.md`, política de suporte, SBOM no build |
| Auto-merge | Ativado para PRs fora das áreas sensíveis, depois dos checks obrigatórios |

### Marco 1: AuthN e AuthZ mínimos, ponta a ponta

**Critério de conclusão:** no laboratório, o cliente de teste A faz login via
OIDC com senha e com passkey numa organização de um tenant, recebe tokens,
consulta uma permissão RBAC pelo AuthZEN e recebe a decisão correta; a própria
plataforma usa esse RBAC nos papéis administrativos; todos os testes de abuso da
seção 11 passam; metas de latência do ADR 0011 medidas por benchmark.

### Camadas seguintes

Ordem e contratos no `docs/roadmap.md`.

## 13. Mudanças em relação ao brief

| Brief | Mudança |
|---|---|
| 4.1 estrutura | Sem `notifications/`; `cmd/iam/` (ADR 0003) |
| 4.2 versão | Versão única (ADR 0008) |
| 4.3 API | Protobuf + Connect; OpenAPI gerada (ADR 0006) |
| 4.4 event sourcing | Não (ADR 0005) |
| 4.15 estágio 0 | Instância compartilhada do laboratório (ADR 0001) |
| 4.17 issuer | Por caminho (ADR 0007) |
| 4.29 motor e linguagem | Motor próprio, núcleo Cedar (ADRs 0014, 0015) |
| 4.30 exceção de segurança | Inclui JOSE (ADR 0010) |
| Seção 8 licença | Apache-2.0 (ADR 0009) |
| Demais [PARA DISCUSSÃO] | Adotados como escritos, salvo onde um ADR diz o contrário |

## 14. Adiado (com o motivo)

| Item | Quando |
|---|---|
| Monetização e recursos pagos | Antes de oferecer versão gerenciada; pontos de extensão já evitam bifurcar o código |
| Nomes definitivos | Antes da v1 (`docs/market-study/naming.md`) |
| Metas de disponibilidade, RPO e RTO por estágio | Ao sair do laboratório (estágio 1) |
| Endereço sob `lab.aleogr.dev`, Cloudflare Access | Quando houver serviço e painel publicados |
| Token do Cloudflare | Não criar enquanto o DNS for manual |
