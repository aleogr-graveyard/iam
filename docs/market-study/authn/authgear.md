# Authgear

| Campo | Valor |
|---|---|
| Categoria | AuthN, serviço self-hosted e SaaS (CIAM, "alternativa open source a Auth0/Clerk/Firebase Auth") |
| Site / repositório | https://www.authgear.com · https://github.com/authgear/authgear-server |
| Licença e modelo de negócio | Apache-2.0 (servidor). Authgear Cloud: Free (MAU ilimitado, 2 apps), Developers US$ 50/mês, Business US$ 500/mês (25k MAU), Enterprise sob consulta (residência de dados, gateway SMS próprio, SLA). Discurso "sem feature gates": todo plano inclui todos os recursos; cobra por apps, assentos de admin, retenção de log e SMS/WhatsApp. **Authgear ONCE**: licença paga uma vez, self-hosted em uma VM, 1 projeto, sem HA, sem WhatsApp, 2 anos de updates |
| Stack | Go (servidor), React (portal), Deno para hooks em JS/TS. PostgreSQL (com `pg_partman`) + Redis. Docker/Helm |
| Maturidade | Repositório criado em jun/2020. ~2,1 mil estrelas, 130 forks, ~360 issues abertas, push diário. Adotantes citados no README: Bupa (HK), CIMIC Group. Comunidade pública pequena (Discord) |
| Acesso | 2026-11-03 |

> Limitação: pouca discussão pública encontrada (HN sem threads relevantes; Reddit inacessível). Fontes principais: docs.authgear.com (inclui `llms.txt`), authgear.com, API do GitHub.

## Resumo
- CIAM orientado a apps de consumo e mobile: SDKs nativos, login biométrico, App2App, usuários anônimos, WhatsApp OTP.
- Fluxos de autenticação customizáveis em YAML (`authentication_flow`) e Authentication Flow API para UI própria.
- Defesa anti-fraude com detecção de **SMS pumping** em modo observar/proteger.
- DPoP para refresh tokens em apps nativos; Auth para MCP com DCR e Client ID Metadata Document.
- Produto e comunidade menores que Keycloak/Zitadel/Logto; poucos advisories (4, todos 2026).

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ◐ | Senha (política, expiração), passwordless por OTP/link mágico (e-mail, SMS, WhatsApp), passkeys, biometria (iOS/Android), social (Google, Apple, Facebook, GitHub, LinkedIn, WeChat), enterprise OIDC (Entra ID, AD FS, Azure AD B2C), LDAP "mínimo" (PRs de 2024). SAML como IdP; SAML como SP (entrada): não encontrado. Certificado/ICP: não encontrado |
| 2 | MFA e step-up | ● | TOTP, SMS, e-mail, WhatsApp, passkeys, códigos de recuperação; re-autenticação para ações sensíveis; "Adaptive MFA" listado no produto (detalhe técnico não verificado) |
| 3 | Ciclo de vida da conta | ◐ | Cadastro (pode desabilitar público), verificação, recuperação, vinculação de contas sociais, exclusão iniciada pelo usuário, desabilitar usuário, período de validade da conta, usuários anônimos promovíveis, import/export. SCIM e convites: não encontrados |
| 4 | Sessões e tokens | ◐ | Sessão por cookie configurável, refresh tokens por app, **DPoP para refresh token** (somente apps nativos; web sem suporte), revogação. Back-channel logout: não encontrado |
| 5 | Protocolos | ◐ | OAuth 2.0/OIDC, SAML IdP, DCR e CIMD (para MCP), WebAuthn. Device flow, CIBA, PAR, FAPI, SCIM: não encontrados no índice de docs |
| 6 | Máquinas e agentes | ◐ | M2M tokens (client credentials), Auth para servidores MCP. Chaves de API e workload identity: não encontrados |
| 7 | Detecção e defesa | ● | Brute force/account lockout, rate limits, bot protection (reCAPTCHA, Turnstile), IP blocklist, Fraud Protection contra SMS pumping. Senhas vazadas: não encontrado |
| 8 | Privacidade e conformidade | ◐ | Exclusão de conta, export de usuários, residência de dados (UK/EU via private cloud/self-host; região EU na nuvem "em breve"), anonimização (issue #2731) |
| 9 | Experiência e administração | ● | Portal admin, AuthUI pré-construída com branding, páginas de configurações do usuário, i18n, UI própria via Authentication Flow API, webhooks e hooks JS/TS (Deno), eventos bloqueantes e não bloqueantes, Admin API GraphQL, audit log (inclui ações de admin) |
| 10 | Operação e migração | ◐ | Helm, Docker Compose, arquiteturas de referência AWS/Azure/GCP/on-prem, guias de migração bulk/rolling/zero-downtime, import de usuários |
| 11 | Identidade descentralizada | ○ | Não encontrado |

## Diferenciais / exclusivos
- **Fraud Protection contra SMS pumping**: sinais por IP, por país do telefone e países por IP; modo "observe" padrão antes de "protect"; decisões com códigos de motivo. Alinha-se ao "alertar em vez de bloquear" (premissa 4).
- **Authgear ONCE**: modelo de licença perpétua para self-host simples, em contraste com cobrança por MAU.
- **Fluxos em YAML** (`signup_flow`, `login_flow`, `signup_login_flow`, passos `identify`/`verify`/`authenticate`/`create_authenticator`) com hooks bloqueantes entre passos.
- **Mobile-first**: App2App login, biometria, usuários anônimos, pre-authenticated URLs (SSO de app nativo para web).
- **OTP test mode** para testes e revisão de lojas de apps.

## O que a comunidade elogia (com evidência)
- Não encontrado em fontes independentes (HN e GitHub com poucas reações; issue mais reagida tem 3 reações). Busca em HN Algolia por "authgear" sem resultados relevantes.
- Material de marketing destaca preço com MAU ilimitado e comparação com Auth0/Frontegg (viés: fornecedor, comparativo próprio).

## Do que a comunidade se queixa (com evidência)
- Evidência pública escassa. Issues abertas antigas sugerem lacunas: login do portal via ADFS/SAML (#3607, 2024), suporte a JWT em cookies (#1180, 2021), testes do token handler (#3485).
- LDAP aparece como especificação aberta (#4485) com "Minimal LDAP support" entregue (#4627): suporte limitado.

## Histórico de segurança

| ID | Ano | Classe | Lição |
|---|---|---|---|
| GHSA-c7wp-6c82-59rm (alto) | 2026 | Path traversal em `/static` expõe config do tenant e código de hooks a anônimos | Servir estáticos isolados de configuração; testes de traversal no CI |
| GHSA-f439-m4xv-94mf (alto) | 2026 | Upload de imagem servido como documento ativo | Uploads com `Content-Type` forçado, `Content-Disposition` e domínio separado |
| GHSA-fxww-5prg-vqgf (alto) | 2026 | SSRF via URLs de webhook e OIDC configuradas pelo projeto | Webhooks e conectores de saída com allowlist/bloqueio de redes internas |
| GHSA-cwgq-fghf-4473 (médio) | 2026 | Proxy da Admin API sem controle de acesso em GET | Autorização não pode depender do verbo HTTP |

Nenhum CVE atribuído aos quatro. Classes (SSRF, traversal, upload) são de plataforma multi-tenant, não de protocolo.

## Arquitetura e decisões notáveis
- Servidor Go único com portal separado; Postgres particionado (`pg_partman`, provavelmente para logs/auditoria) e Redis (sessões, rate limit).
- Multi-projeto (cada projeto = tenant com config YAML própria `authgear.yaml` + `authgear.secrets.yaml`), editável no portal ("Edit Config").
- Hooks em Deno executam código do cliente no servidor; um advisory expôs o código desses hooks (classe de risco a observar).
- HA de referência: 3 nós K8s, Postgres primário/standby com `pg_auto_failover`, HAProxy.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Multi-tenancy no nível de **projeto** (pool de usuários isolado). Organizações B2B dentro de um projeto: não encontrado no índice de docs.
- Extensibilidade: webhooks, hooks JS/TS, gateway SMS customizado (webhook/script), provedor de e-mail customizado, Admin API GraphQL.
- Banco: somente PostgreSQL + Redis.
- Soberania: self-host, private cloud, residência UK/EU como argumento comercial.

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **não** (não encontrado).
- **Autorização sensível ao estado / transições**: **não**. Autorização limitada a roles e groups no token/UserInfo.

## Implicações para o projeto
- **Adotar**: proteção contra SMS pumping com modo observar antes de bloquear e motivos registrados (premissa 4; restrição de SMS do brief 5.1).
- **Adotar**: OTP test mode configurável por tenant para testes e revisão de lojas, auditado.
- **Adotar**: DPoP em refresh tokens para clientes nativos como primeiro passo (chave em armazenamento seguro do dispositivo).
- **Evitar**: hooks com código arbitrário sem isolamento forte; SSRF em URLs configuradas por tenant precisa de controle central.
- **Oportunidade**: combinar fluxos declarativos (como o YAML do Authgear) com UI visual e testes automatizados de fluxo.
- **Oportunidade**: modelo de licença/preço não baseado em MAU (ONCE) como referência para monetização (brief seção 8).

## Fontes
- https://github.com/authgear/authgear-server (README via raw.githubusercontent.com) — 2026-11-03
- https://api.github.com/repos/authgear/authgear-server — 2026-11-03
- https://api.github.com/repos/authgear/authgear-server/security-advisories — 2026-11-03
- https://api.github.com/search/issues?q=repo:authgear/authgear-server+is:issue&sort=reactions — 2026-11-03
- https://api.github.com/search/issues?q=repo:authgear/authgear-server+ldap — 2026-11-03
- https://www.authgear.com/pricing — 2026-11-03 — viés: fornecedor (inclui comparativo com concorrentes)
- https://docs.authgear.com/llms.txt — 2026-11-03
- https://docs.authgear.com/authgear-once/what-is-authgear-once — 2026-11-03
- https://docs.authgear.com/security/sender-constraining — 2026-11-03
- https://docs.authgear.com/security/fraud-protection — 2026-11-03
- https://docs.authgear.com/security/bot-protection — 2026-11-03
- https://docs.authgear.com/authentication-and-access/authentication/custom-authentication-flow — 2026-11-03
- https://docs.authgear.com/authentication-and-access/sessions — 2026-11-03
- https://docs.authgear.com/admin-and-operations/user-management/manage-users-roles-and-groups — 2026-11-03
- https://docs.authgear.com/deployment/local — 2026-11-03
- https://docs.authgear.com/deployment/architecture-guides/on-premises-reference-architecture — 2026-11-03
- https://hn.algolia.com/api/v1/search?query=authgear — 2026-11-03 (sem resultados relevantes)
