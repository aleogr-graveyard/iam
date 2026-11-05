# Pomerium (Core, Zero e Enterprise)

| Campo | Valor |
|---|---|
| Categoria | AuthN/AuthZ na borda: identity- and context-aware access proxy (zero trust, "clientless"), com acesso SSH, TCP/UDP, Kubernetes e MCP. Não é IdP |
| Site / repositório | https://www.pomerium.com · https://github.com/pomerium/pomerium |
| Licença e modelo de negócio | **Código-fonte Apache-2.0.** **Mudança em 2026**: binários, pacotes e imagens oficiais lançados após 01/nov/2026 passam à "Pomerium Community Edition License 1.0" (Apache modificada): gratuitos só para quem tem < US$ 10 mi de receita bruta **e** < 100 pessoas; demais compram licença ou compilam do fonte; coleta de uso ("Usage Data") ligada por padrão com opt-out; rescisão por violação. Planos: **Zero Personal** (grátis: plano de controle gerenciado, 1 cluster, 10 rotas, 5 políticas), **Zero Business** (US$ 9/usuário/mês ou US$ 7 anual: 5 clusters, 100 rotas/políticas, logs de acesso, histórico), **Enterprise** (sob consulta: plano de controle self-hosted, sincronização de diretório, dados externos, Rego, administração delegada, herança de políticas, gestão de sessões, gravação SSH, whitelabel) |
| Stack | Go 1.27 + Envoy (build customizado `pomerium/envoy-custom`, configurado via go-control-plane), OPA/Rego como motor sob a PPL, pgx (PostgreSQL), WebAuthn próprio (`pomerium/webauthn`), OpenTelemetry |
| Maturidade | ~5,1 mil estrelas, 368 forks, 91 issues abertas. Doc "latest stable" v0.32; advisory de out/2026 corrigido em 0.33.5 (ainda 0.x). Empresa com clientes Enterprise (página Customers); idade do repositório não verificada (API do GitHub bloqueada) |
| Acesso | 2026-11-03 |

> Limitações: `api.github.com` bloqueado; `www.pomerium.com` lido via curl. Orçamento de buscas web esgotado no fim da pesquisa (sem reação da comunidade à mudança de licença de nov/2026).

## Resumo
- Proxy de acesso com autorização contínua (cada requisição reavaliada), política declarativa (PPL, sobre OPA) e repasse de identidade assinada (JWT próprio) ao upstream.
- Quatro serviços lógicos (proxy, authenticate, authorize, databroker) que rodam juntos num binário ou separados.
- **Pomerium Zero**: plano de controle SaaS (console, API, Terraform) + plano de dados no ambiente do cliente; o tráfego não passa pela Pomerium.
- Aposta recente forte em **agentes de IA/MCP**: autenticação de servidores MCP, controle por ferramenta (`mcp_tool`), gestão de tokens OAuth upstream.
- Mudança de licença dos binários oficiais (nov/2026) é o fato de negócio mais relevante: fonte continua Apache, mas a distribuição oficial vira "source-available" condicional.

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ◐ | Delegada a IdP OIDC/OAuth (Apple, Auth0, Entra ID, Cognito, Descope, GitHub, GitLab, Google, Keycloak, Okta, OneLogin, Ping, OIDC genérico). "Hosted Authenticate Service" da Pomerium dispensa IdP próprio. Aceita access/ID token do IdP direto ("Direct IdP Token Authentication"). Sem senha/passkey/SAML próprios |
| 2 | MFA e step-up | ◐ | MFA delegado ao IdP. **Device identity** com WebAuthn/hardware (TPM, chaves) como fator de autorização por rota |
| 3 | Ciclo de vida da conta | ◐ | Não gere contas. Directory sync (Enterprise) para grupos/usuários usados em política |
| 4 | Sessões e tokens | ◐ | Sessão no databroker (cookie só referencia), claims do IdP atualizados periodicamente via UserInfo (mudanças no IdP valem sem novo login). Gestão/revogação central de sessões: só Enterprise. Tokens do IdP nunca vão ao cliente nem ao upstream; upstream recebe JWT Pomerium assinado |
| 5 | Protocolos | ◐ | RP OIDC/OAuth2; mTLS cliente e upstream; atua como servidor de autorização OAuth para clientes MCP. Sem SAML, SCIM |
| 6 | Máquinas e agentes | ● | Contas de serviço (Enterprise/Zero), tokens do IdP para acesso programático, acesso MCP com política por ferramenta e refresh automático de tokens upstream, SSH nativo, TCP/UDP via CLI/Desktop |
| 7 | Detecção e defesa | ◐ | Contexto de dispositivo e de requisição em política; postura de dispositivo via FleetDM (Enterprise). Rate limiting nativo: não encontrado |
| 8 | Privacidade e conformidade | ◐ | Logs de acesso/autorização e auditoria administrativa; dados ficam no plano de dados do cliente. Databroker não cifra valores em repouso (recomenda cifrar o volume/Postgres) |
| 9 | Experiência e administração | ◐ | Console (Zero gerenciado / Enterprise self-hosted), editor visual de políticas, API de gestão, provedor Terraform, Ingress Controller K8s. Whitelabel só Enterprise. Página 403 customizável (#1333) |
| 10 | Operação e migração | ◐ | Binário único ou serviços separados; databroker em memória, arquivo ou PostgreSQL (obrigatório com várias réplicas); OpenTelemetry; health checks |
| 11 | Identidade descentralizada | ○ | Não |

## Diferenciais / exclusivos
- **Autorização contínua**: política reavaliada em toda requisição, inclusive dentro de sessão existente, com claims atualizados do IdP.
- **PPL** (YAML declarativo: `allow`/`deny` com `and`/`or`/`not`, critérios `email`, `groups`, `claim`, `device`, `date`, `http_method`...) compilada para Rego; Rego direto só no Enterprise.
- **Separação plano de controle × plano de dados** (Zero): SaaS gerencia configuração; tráfego e sessões ficam no cliente.
- **Hosted Authenticate Service**: experimentar sem configurar IdP.
- **JIT por política**: critério `date` com `before`/`after` expira acesso sozinho; app de exemplo `jit-example` faz solicitação/aprovação via API do Zero (guia JIT).
- **MCP de primeira classe**: proteger servidor MCP, limitar ferramentas por usuário/grupo, delegar a LLM, "tunnel to ChatGPT".
- **Gravação de sessões SSH** com auditoria (Enterprise).

## O que a comunidade elogia (com evidência)
- Opção open source madura para "BeyondCorp" self-hosted; usado em homelab sobre Envoy ([HN 50051106](https://news.ycombinator.com/item?id=50051106), out/2026).
- Integrações citadas em discussões (Tailscale + Pomerium, Traefik + Pomerium) ([HN 35977675](https://news.ycombinator.com/item?id=35977675), [HN 23078869](https://news.ycombinator.com/item?id=23078869)).
- Issues mais votadas são quase todas **fechadas** (11 de 12): curinga de vhost (#1677), bypass por IP (#1943), repasse de access token (#1915), JWT em cookie (#1282) ([GitHub](https://github.com/pomerium/pomerium/issues?q=is%3Aissue+sort%3Areactions-%2B1-desc)). Indica boa absorção de demandas.
- Engenharia publicada sobre Envoy (bug use-after-free de DNS, health checks) ([HN 46247564](https://news.ycombinator.com/item?id=46247564), [HN 45761956](https://news.ycombinator.com/item?id=45761956)).

## Do que a comunidade se queixa (com evidência)
- **Recursos essenciais no pago**: gestão/revogação de sessões, sync de diretório, Rego, administração delegada e whitelabel só no Enterprise (página de preços).
- **Licença dos binários oficiais (nov/2026)**: limite de receita/pessoal, telemetria por padrão e foro na Califórnia (texto da licença). Reação pública: não encontrada (busca esgotada).
- **Forward auth**: pedido de reconsiderar o suporte a forward-auth (#5567, fechado em out/2025); regex em políticas pedido desde 2020 (#659, aberta).
- **Dependência de Envoy**: peso e superfície herdados (advisories de Envoy embutido, GHSA-j34v-3552-5r7j).
- Pouca tração no HN (histórias com ≤ 10 pontos), sinal de comunidade menor que oauth2-proxy.

## Histórico de segurança
~12 advisories ([lista](https://github.com/pomerium/pomerium/security/advisories)). Destaques:

| ID | Ano | Severidade | Classe |
|---|---|---|---|
| GHSA-cqmh-8f5g-9m2w | 2026 | Moderada (6.3) | Bypass de `mcp_tool`: authorize via só 4096 bytes do corpo, não parseava lote JSON-RPC e `encoding/json` casa chaves sem diferenciar maiúsculas (`name` × `Name`) — o upstream executava outra ferramenta. Corrigido em 0.33.5 |
| GHSA-ggw3-5987-rx77 | 2026 | Alta | Exaustão de memória pré-auth (descompressão zstd sem limite) |
| GHSA-pj2c-mjp8-p2m9 | 2026 | Baixa | DoS por reuso de conexão upstream |
| GHSA-r7rh-jww5-5fjr | 2024 | Moderada | Escopo indevido de token de conta de serviço na API do databroker |
| GHSA-rrqr-7w59-637v | 2024 | Moderada | Tokens expostos na resposta de user info |
| GHSA-pvrc-wvj2-f59p / CVE-2023-33189 | 2023 | Crítica (10.0) | Decisão de autorização incorreta com requisições forjadas (CWE-285) |
| GHSA-q98f-2x4p-prjr | 2022 | Moderada | Endpoints de debug/métricas expostos |
| GHSA-j6wp-3859-vxfg | 2021 | Moderada | Claims OIDC não atualizados (dados obsoletos na autorização) |
| GHSA-gjcg-vrxg-xmgv | 2021 | Alta | Frames HTTP/2 GOAWAY/SETTINGS |

Lições para nós:
- **O PDP precisa ver exatamente o que o upstream executa**: truncamento de corpo, lotes e parsing JSON permissivo (case-insensitive, chaves duplicadas) criam bypass. Em Go, não usar `encoding/json` padrão para decisões de segurança sem rejeitar chaves duplicadas/variantes de caixa; falhar fechado quando o corpo não couber.
- **Limites em toda descompressão e todo parsing pré-auth**.
- **Tokens de serviço com escopo mínimo** e APIs internas (databroker) também autorizadas.
- **Claims obsoletos são falha de autorização**: definir e testar política de atualização de atributos.
- **Endpoints de debug/métricas** fora da interface pública por padrão.

## Arquitetura e decisões notáveis
- **Serviços**: Proxy (Envoy) → Authorize (gRPC, avalia política, monta cabeçalhos) → Authenticate (fluxo OIDC, cria sessão) → Databroker (KV de "records" versionados com leases, consulta e streaming de mudanças; caches locais nos serviços quentes) (doc Architecture, Data storage).
- Sessão: no começo tudo em cookie; migraram para databroker por limites de cookie/cabeçalho, SPAs quebrando, falta de visibilidade central e grupos fixos no login (doc "Persistence & Data storage" — lição direta para nós).
- Tokens do IdP nunca saem do databroker; upstream recebe JWT Pomerium assinado (`X-Pomerium-Jwt-Assertion`) com SDKs de verificação (Go, JS).
- Configuração estática (arquivo/env) no Core; no Zero/Enterprise/Ingress, a própria configuração vive no databroker.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- **Multi-tenancy**: namespaces com papéis e herança de políticas (Enterprise); vários clusters por organização (Zero: 1/5; Enterprise: multi-cluster).
- **Extensibilidade**: databroker pode ser substituído implementando a interface gRPC; dados externos e Rego no Enterprise; SDKs Go/JS.
- **Configuração**: YAML/env, CRDs do Ingress Controller, API de gestão, Terraform.
- **Bancos**: memória (padrão), arquivo, PostgreSQL. Sem cifra própria em repouso.
- **Soberania**: plano de dados sempre no ambiente do cliente (também no Zero); o plano de controle do Zero é SaaS da Pomerium.

## Verificação de diferenciais candidatos do projeto
- Aprovação multipartes (M de N): **não**. JIT por janela de tempo (`date`) e app de exemplo com aprovação simples via API do Zero; sem quórum.
- Autorização sensível ao estado / transições declarativas: **não**. Política avalia identidade, dispositivo, requisição e tempo; nenhum conceito de estado do recurso ou transição.

## Implicações para o projeto
1. **Adotar** sessão do lado do servidor com cache local quente (padrão databroker) e atualização periódica de atributos do IdP; decisão a cada requisição no modo proxy.
2. **Adotar** separação plano de controle × plano de dados como opção de produto gerenciado: atende soberania (4.8) e monetização (seção 8) sem trafegar dados do cliente.
3. **Adotar** JWT próprio e assinado para o upstream, nunca repassar tokens do IdP; publicar SDKs de verificação.
4. **Evitar** gargalos de licença: a mudança de nov/2026 mostra o risco percebido de "open core" com binário condicionado; deixar claro desde o início o que é aberto (seção 8).
5. **Evitar** a falha `mcp_tool`: o PDP deve analisar o corpo inteiro com parser estrito ou negar; incluir testes de "parser differential" no CI ofensivo.
6. **Oportunidade**: gestão/revogação de sessões e sync de diretório são pagos no Pomerium; tê-los no núcleo aberto diferencia.
7. **Oportunidade**: JIT com aprovação M de N nativa na borda (Pomerium só oferece janela de tempo + app de exemplo).

## Fontes
- https://github.com/pomerium/pomerium — 2026-11-03
- https://raw.githubusercontent.com/pomerium/pomerium/main/README.md — 2026-11-03
- https://raw.githubusercontent.com/pomerium/pomerium/main/licenses/POMERIUM-COMMUNITY-LICENSE-1.0.txt — 2026-11-03
- https://raw.githubusercontent.com/pomerium/pomerium/main/go.mod — 2026-11-03
- https://github.com/pomerium/pomerium/security/advisories — 2026-11-03
- https://github.com/pomerium/pomerium/security/advisories/GHSA-cqmh-8f5g-9m2w — 2026-11-03
- https://github.com/advisories/GHSA-pvrc-wvj2-f59p — 2026-11-03
- https://github.com/pomerium/pomerium/issues?q=is%3Aissue+sort%3Areactions-%2B1-desc — 2026-11-03
- https://www.pomerium.com/pricing — 2026-11-03 — fornecedor
- https://www.pomerium.com/docs/internals/architecture — 2026-11-03
- https://www.pomerium.com/docs/internals/data-storage — 2026-11-03
- https://www.pomerium.com/docs/internals/sessions — 2026-11-03
- https://www.pomerium.com/docs/reference/databroker — 2026-11-03
- https://www.pomerium.com/docs/capabilities/hosted-authenticate-service — 2026-11-03
- https://www.pomerium.com/docs/capabilities/device-identity — 2026-11-03
- https://www.pomerium.com/docs/guides/jit — 2026-11-03
- https://www.pomerium.com/docs/integrations/user-identity/identity-providers — 2026-11-03
- https://www.pomerium.com/docs/sitemap.xml (lista de capacidades MCP, SSH, gravação de sessão) — 2026-11-03
- HN via API Algolia: 50051106, 35977675, 23078869, 46247564, 45761956 — 2026-11-03
