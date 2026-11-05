# Authentik

| Campo | Valor |
|---|---|
| Categoria | AuthN, serviço self-hosted (IdP completo + proxy/forward auth via outposts) |
| Site / repositório | https://goauthentik.io · https://github.com/goauthentik/authentik |
| Licença e modelo de negócio | Open core. Repositório principal MIT; código em `authentik/enterprise/` sob a "authentik EE License" (source-available). Enterprise US$ 5/usuário interno/mês e US$ 0,02/usuário externo/mês; Enterprise Plus a partir de US$ 20 mil/ano (FIPS, multi-instância, SLA). Sem SaaS gerenciado. Promessa pública: nenhum recurso migra do OSS para o enterprise |
| Stack | Python/Django (core, API, flow executor) + Go (servidor/roteador e outposts: proxy, LDAP, RADIUS, RAC). UI em Lit (web components). Somente PostgreSQL: Redis removido em 2025.10 (tarefas, cache, sessão do outpost embutido e WebSocket migrados para o Postgres, com ~50% mais conexões) |
| Maturidade | Repositório criado em dez/2019. ~26,2 mil estrelas, ~2,1 mil forks, ~1,2 mil issues abertas, push diário. Versão atual 2026.8 (calendário AAAA.M, ~3 releases/ano). "OpenID Certified" (set/2026). Empresa declara 1M+ instalações únicas; CEO afirma que a Cloudflare usa o authentik para identidade de colaboradores |
| Acesso | 2026-11-03 |

> Limitação: páginas do GitHub (web) e parte da API bloqueadas por rate limit na sessão. Usei a API do GitHub (metadados, advisories, busca de issues), docs.goauthentik.io, goauthentik.io e a API Algolia do HN. Reddit inacessível.

## Resumo
- IdP generalista com motor de **flows → stages → policies → bindings**: todo processo (login, cadastro, recuperação, autorização de app, logout, solicitação de acesso) é um fluxo configurável pela UI, API, Terraform ou blueprints YAML.
- Protocolos amplos: OIDC/OAuth2, SAML, LDAP, RADIUS, Kerberos, SCIM, WS-Fed, SSF, proxy reverso e RDP/SSH/VNC no navegador (RAC).
- Popular em homelab e PMEs pela UI administrativa; queixas sobre peso (Python), complexidade de proxy/outposts e regressões em upgrades.
- Volume muito alto de advisories, concentrados em 2026 (SAML, forward auth, controle de acesso da API, bypass de stages).

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ● | Senha, passkeys/WebAuthn sem senha, sources OAuth/OIDC/SAML (Apple, Google, Microsoft, Discord, Twitch, etc.), LDAP/AD e FreeIPA (sync), Kerberos/SPNEGO, SCIM source. mTLS como stage (Enterprise): base técnica para certificado, sem perfil ICP-Brasil. Link mágico via Email stage |
| 2 | MFA e step-up | ● | TOTP, WebAuthn, Duo (push), SMS (Twilio/genérico), e-mail, códigos estáticos; Authenticator Validation stage. Step-up via authorization flow por provider. "Adaptativo" via policies (reputação, GeoIP, expressão Python), não motor de risco dedicado |
| 3 | Ciclo de vida da conta | ● | Enrollment, invitation, recovery, unenrollment (User delete stage), vinculação de source a usuário, JIT por source, SCIM provider (saída) e source (entrada), Account Lockdown stage (2026.5, Enterprise), offboarding agendado (Enterprise). Verificação de idade: não encontrado |
| 4 | Sessões e tokens | ◐ | Revogação de sessões, front/back-channel logout e Single Logout, refresh tokens, introspecção/revogação. DPoP apenas para **ID token key-bound** (`bound_key`), não para access token. Pedido de sessão "deslizante" aberto (issue #14304) |
| 5 | Protocolos | ◐ | OIDC (certificado), OAuth2 com device code e token exchange (RFC 8693), DCR, SAML 2.0, SCIM, WebAuthn, WS-Fed, LDAP, RADIUS, SSF (Enterprise). PAR, CIBA, FAPI: não encontrados na doc do provider OAuth2 |
| 6 | Máquinas e agentes | ● | Service accounts, tokens/app passwords, client credentials, "agent accounts" para identidade não humana (Enterprise), token exchange com delegação |
| 7 | Detecção e defesa | ◐ | Captcha stage, Reputation policy, GeoIP, Password policy com HIBP e zxcvbn. Sem proteção específica contra bots/SMS pumping; rate limiting delegado ao proxy |
| 8 | Privacidade e conformidade | ◐ | Consent stage, User delete stage, retenção de eventos, auditoria ampliada e relatórios CSV (Enterprise), FIPS (Enterprise Plus). Exportação de dados pelo titular: não encontrado |
| 9 | Experiência e administração | ● | Admin UI completa, user dashboard, brands (domínio/tema por marca), layouts de flow, i18n via Transifex, notificações por transports (webhook, e-mail), eventos |
| 10 | Operação e migração | ◐ | Blueprints YAML declarativos, provider Terraform, Helm, CloudFormation. Operator Kubernetes é o pedido mais votado (issue #5675, aberto desde 2023). Migração de outros IdPs só por guias |
| 11 | Identidade descentralizada | ○ | Não encontrado |

## Diferenciais / exclusivos
- **Flow engine visual**: designações (authentication, authorization, enrollment, recovery, invalidation, stage configuration, unenrollment), ~25 tipos de stage e políticas avaliadas na planificação ou dinamicamente antes de cada stage.
- **Access requests (2026.8, preview, Enterprise)**: o usuário solicita acesso a uma aplicação, revisores aprovam com **mínimo de revisores** (total ou por grupo), concessão com expiração automática e auditoria. Ver verificação abaixo.
- **Outposts** (Go) que levam o IdP até a borda: proxy/forward auth para Traefik, Caddy, nginx; servidor LDAP; RADIUS (inclusive EAP-TLS no Enterprise); RAC para RDP/SSH/VNC no navegador.
- Device trust (Chrome Enterprise) e "endpoint devices"/agente para login em Windows/SSH (preview).
- Integrações de saída com Google Workspace e Entra ID (Enterprise).

## O que a comunidade elogia (com evidência)
- UI administrativa que poupa configuração manual frente ao Authelia ("8h no Authelia… authentik tem UI de gestão e é simples de configurar") — HN 40953205 (2024).
- Preferido ao Keycloak em homelab por facilidade com Docker — HN 36385695 (2023).
- Adoção corporativa citada pelo próprio CEO (Cloudflare) — HN 48676777 (2026) (viés: fornecedor).
- Uso com OPKSSH, Tailscale/Headscale e Kubernetes recorrente em comentários — HN 49145284, 47915088.

## Do que a comunidade se queixa (com evidência)
- **Peso e desempenho de Python/Django**: "usa bastante RAM" — HN 31261787; a própria empresa publicou "Django: 500% mais lento para rodar, 200% mais rápido para construir" — HN 35186568 (viés: fornecedor, assume o trade-off).
- **Complexidade**: "grande, complicado e modular; a complexidade aparece no docker-compose" — HN 31409471; "exagero para homelab" — HN 50151718.
- **Proxy/outposts e upgrades**: issues mais reagidas são outpost não saudável após 2025.12.2 (#19878, 76 reações), redirecionamentos aleatórios do proxy (#6886), `/outpost.goauthentik.io` não encontrado (#8956), forward auth com HAProxy (#5768).
- Logout do authentik não invalidava sessões de apps protegidos (#2023, 48 reações), sessões que expiram em dispositivos usados com frequência (#14304).
- Dependência de container: pedido de instalação sem Docker/K8s (#4273).

## Histórico de segurança
~60 advisories publicados no GitHub; ~35 só em 2026. Seleção:

| ID | Ano | Classe | Lição |
|---|---|---|---|
| CVE-2024-47070 (crítico) | 2024 | Bypass de senha via `X-Forwarded-For` | Nunca confiar em cabeçalhos de proxy sem lista de proxies confiáveis; ver também CVE-2023-36456 |
| CVE-2023-46249 (crítico) | 2023 | Tomada da instalação ao excluir admin padrão (setup flow reabria) | Bootstrap/usuário master deve ser idempotente e não reabrível (brief 4.23) |
| CVE-2023-26481 (crítico) | 2023 | Checagem insuficiente de usuário no token do Email stage | Tokens de fluxo devem ser ligados ao usuário e à transação |
| CVE-2026-49448 (crítico) | 2026 | Bypass de SourceStage com POST vazio | Motor de fluxos: cada stage precisa de "fail closed" para entrada vazia/inesperada |
| CVE-2026-25227 (crítico) | 2026 | RCE autenticado pelo endpoint de teste de Policy/Property Mapping (Python) | Políticas como código arbitrário = superfície de RCE; preferir linguagem restrita e sandbox |
| CVE-2026-42849 (crítico) | 2026 | XSS refletido na SFE | Interface de fluxo alternativa recebe a mesma cobertura (brief 5.1, fluxos pouco usados) |
| CVE-2026-25922, -40165, -57580, -47201, -94612, -41577 | 2026 | SAML source: bypass de assinatura, XSW, injeção de comentário no NameID, replay, ausência de validação de Conditions/audience | SAML é a maior fonte de falhas; usar biblioteca madura, testes de conformidade negativos e fuzzing |
| CVE-2026-25748, GHSA-5wcc-hf24-rf5h | 2026 | Bypass de forward auth (cookie malformado; `X-Original-URI` controlado pelo cliente) | Modo proxy exige canonicalização e contrato estrito com o proxy |
| CVE-2023-48228, CVE-2024-23647, CVE-2024-52289 | 2023–24 | PKCE insuficiente/downgrade; redirect URI por regex por padrão | PKCE obrigatório e comparação exata de redirect URI por padrão |
| CVE-2024-42490, -37905, -40166, -40172, -94609, -61574 | 2024–26 | Autorização insuficiente na API (tokens, client_secret, superuser, RAC) | API administrativa precisa de autorização por objeto testada sistematicamente |
| CVE-2025-29928, CVE-2025-53942, CVE-2025-64521 | 2025 | Sessões não revogadas; contas inativas/service accounts desativadas autenticando via source/OAuth | Estado da conta checado em todo caminho de emissão de token |
| CVE-2026-94606, CVE-2026-54730 | 2026 | Bypass de MFA por e-mail (override de destinatário); bypass de device trust | Destinos de fator vêm do servidor, nunca do cliente |

## Arquitetura e decisões notáveis
- Contêiner `server` (roteador Go + core Django + outpost embutido) e `worker` (tarefas, notificações); PostgreSQL como único estado. Remoção do Redis (2025.10) simplifica operação ao custo de carga no banco.
- Outposts em Go se conectam ao core por API/WebSocket e podem rodar perto das aplicações (Docker/K8s gerenciados pelo core).
- Escolha declarada de Python pela velocidade de desenvolvimento, aceitando custo de desempenho.
- Flow Planner: políticas avaliadas no planejamento ou antes de cada stage; "denied action" e "policy engine mode" (any/all) por fluxo.
- Blueprints: importação/exportação de configuração inteira em YAML (GitOps); importar blueprint de terceiros é risco (pode conter usuários/políticas).

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- **Brands**: várias marcas/domínios numa instância (antigo "tenants").
- **Tenants** (Enterprise, **alpha** desde 2024.2): um schema PostgreSQL por tenant, criados só via API, licença por tenant; expression policies "têm acesso a todos os tenants" e outpost embutido não suportado.
- Extensibilidade: expression policies e property mappings em Python, custom stages via flows, API REST completa, Terraform.
- Banco: somente PostgreSQL. Soberania: total por ser self-hosted (sem nuvem do fornecedor).

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **parcial**. Access requests (2026.8, preview, Enterprise) têm "Minimum reviewers" com opção por grupo, solicitante não aprova o próprio pedido, expiração automática e eventos auditados. Escopo limitado a acesso a aplicações/entitlements; não é primitivo genérico para ações arbitrárias. Fonte: docs "Access requests".
- **Autorização sensível ao estado / transições declarativas**: **não**. Flows modelam etapas de autenticação, não estados de recursos de negócio.

## Implicações para o projeto
- **Adotar**: composição de fluxos por configuração (flows/stages/policies) como forma de cumprir a premissa 3; com testes de cada stage em "fail closed".
- **Adotar**: configuração declarativa exportável (blueprints) e provider Terraform desde cedo.
- **Evitar**: políticas em linguagem de uso geral (Python) executadas no servidor; usar CEL ou similar com sandbox e limites.
- **Evitar**: confiar em cabeçalhos de proxy; o modo proxy precisa de contrato explícito e testes ofensivos (brief 4.19).
- **Evitar**: SAML implementado em casa sem suíte de testes negativos; priorizar bateria de XSW/comentários/replay.
- **Oportunidade**: multi-tenancy de primeira classe (o authentik ainda está em alpha e pago).
- **Oportunidade**: M de N genérico (além de acesso a apps) e integrado ao AuthZ.
- **Oportunidade**: desempenho previsível (Go) como contraste explícito ao Django.

## Fontes
- https://github.com/goauthentik/authentik — 2026-11-03
- https://api.github.com/repos/goauthentik/authentik — 2026-11-03
- https://api.github.com/repos/goauthentik/authentik/security-advisories — 2026-11-03
- https://api.github.com/search/issues?q=repo:goauthentik/authentik+is:issue&sort=reactions — 2026-11-03
- https://goauthentik.io/pricing/ — 2026-11-03 — viés: fornecedor
- https://docs.goauthentik.io/add-secure-apps/flows-stages/flow/ — 2026-11-03
- https://docs.goauthentik.io/add-secure-apps/flows-stages/stages/ — 2026-11-03
- https://docs.goauthentik.io/add-secure-apps/flows-stages/stages/account_lockdown/ — 2026-11-03
- https://docs.goauthentik.io/add-secure-apps/access-requests/ — 2026-11-03
- https://docs.goauthentik.io/add-secure-apps/providers/oauth2/ — 2026-11-03
- https://docs.goauthentik.io/users-sources/sources/ — 2026-11-03
- https://docs.goauthentik.io/customize/policies/ — 2026-11-03
- https://docs.goauthentik.io/sys-mgmt/tenancy/ — 2026-11-03
- https://docs.goauthentik.io/core/architecture/ — 2026-11-03
- https://docs.goauthentik.io/releases/2025.10/ — 2026-11-03
- https://news.ycombinator.com/item?id=40953205 , 36385695, 31261787, 31409471, 50151718, 48676777 (viés: CEO), 47604611 (vaga, viés: fornecedor), 35186568 (viés: fornecedor) — 2026-11-03
