# Matriz de funcionalidades — AuthN

Acesso: 2026-11-03 a 2026-11-05. Legenda: ● completo, ◐ parcial, ○ ausente, ? não verificado.
Avaliação qualitativa de cada agente de pesquisa; detalhes e fontes no documento de cada player.

Domínios (brief, seção 5.1):
1. Autenticação primária · 2. MFA e step-up · 3. Ciclo de vida da conta · 4. Sessões e tokens ·
5. Protocolos · 6. Máquinas e agentes · 7. Detecção e defesa · 8. Privacidade e conformidade ·
9. Experiência e administração · 10. Operação e migração · 11. Identidade descentralizada

| Player | Categoria | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| [Keycloak](authn/keycloak.md) | Self-hosted | ◐ | ◐ | ◐ | ● | ● | ◐ | ◐ | ◐ | ◐ | ● | ◐ |
| [Zitadel](authn/zitadel.md) | Self-hosted | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ● | ◐ | ○ |
| [Ory](authn/ory.md) | Self-hosted | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ○ |
| [Authentik](authn/authentik.md) | Self-hosted | ● | ● | ● | ◐ | ◐ | ● | ◐ | ◐ | ● | ◐ | ○ |
| [Authgear](authn/authgear.md) | Self-hosted | ◐ | ● | ◐ | ◐ | ◐ | ◐ | ● | ◐ | ● | ◐ | ○ |
| [Logto](authn/logto.md) | Self-hosted | ◐ | ◐ | ◐ | ◐ | ◐ | ● | ◐ | ◐ | ● | ◐ | ○ |
| [Authelia](authn/authelia.md) | Self-hosted | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ○ | ◐ | ○ |
| [Kanidm](authn/kanidm.md) | Self-hosted | ◐ | ◐ | ◐ | ◐ | ◐ | ● | ◐ | ◐ | ◐ | ◐ | ○ |
| [Auth0/Okta](authn/auth0-okta.md) | IDaaS | ◐ | ● | ● | ◐ | ● | ● | ● | ◐ | ● | ◐ | ○ |
| [Entra External ID](authn/entra-external-id.md) | IDaaS | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ |
| [Cognito](authn/cognito.md) | IDaaS | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ○ |
| [FusionAuth](authn/fusionauth.md) | IDaaS / self-hosted | ◐ | ◐ | ● | ◐ | ◐ | ◐ | ◐ | ● | ● | ● | ○ |
| [Clerk](authn/clerk.md) | IDaaS | ● | ◐ | ● | ◐ | ◐ | ◐ | ● | ◐ | ● | ◐ | ○ |
| [WorkOS](authn/workos.md) | IDaaS | ● | ◐ | ● | ◐ | ◐ | ◐ | ● | ◐ | ● | ◐ | ○ |
| [Descope](authn/descope.md) | IDaaS | ● | ● | ● | ● | ● | ● | ◐ | ● | ● | ◐ | ? |
| [Stytch](authn/stytch.md) | IDaaS | ● | ◐ | ● | ◐ | ◐ | ◐ | ● | ◐ | ● | ◐ | ○ |
| [Frontegg](authn/frontegg.md) | IDaaS | ● | ● | ● | ◐ | ◐ | ◐ | ● | ◐ | ● | ◐ | ○ |
| [Better Auth](authn/better-auth.md) | Biblioteca | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ○ |
| [SuperTokens](authn/supertokens.md) | Biblioteca + serviço | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ? | ◐ | ◐ | ○ |
| [Django Auth](authn/framework-auth.md) | Framework | ◐ | ○ | ◐ | ◐ | ○ | ○ | ○ | ○ | ◐ | ◐ | ○ |
| [Devise](authn/framework-auth.md) | Framework | ◐ | ○ | ◐ | ◐ | ○ | ○ | ◐ | ○ | ◐ | ◐ | ○ |
| [ASP.NET Identity](authn/framework-auth.md) | Framework | ◐ | ◐ | ◐ | ◐ | ◐ | ○ | ◐ | ◐ | ◐ | ◐ | ○ |
| [Spring Security 7](authn/framework-auth.md) | Framework | ● | ◐ | ○ | ● | ● | ◐ | ○ | ○ | ○ | ◐ | ○ |
| [Libs Go/Rust](authn/go-rust-libraries.md) | Peças | ◐ | ◐ | ○ | ◐ | ◐ | ◐ | ○ | ○ | ○ | ○ | ○ |
| [oauth2-proxy](authn/oauth2-proxy.md) | Proxy | ◐ | ○ | ○ | ◐ | ◐ | ◐ | ○ | ○ | ◐ | ◐ | ○ |
| [Pomerium](authn/pomerium.md) | Proxy | ◐ | ◐ | ◐ | ◐ | ◐ | ● | ◐ | ◐ | ◐ | ◐ | ○ |
| [Cloudflare Access](authn/cloudflare-access-google-iap.md) | Proxy gerenciado | ◐ | ◐ | ◐ | ◐ | ◐ | ● | ◐ | ◐ | ◐ | ◐ | ○ |
| [Google IAP](authn/cloudflare-access-google-iap.md) | Proxy gerenciado | ◐ | ◐ | ○ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ○ |
| [Supabase Auth](authn/baas.md) | BaaS | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ○ |
| [Firebase / Identity Platform](authn/baas.md) | BaaS | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ○ |
| [Appwrite](authn/baas.md) | BaaS | ◐ | ◐ | ● | ◐ | ◐ | ◐ | ◐ | ◐ | ● | ◐ | ○ |
| [PocketBase](authn/baas.md) | BaaS | ◐ | ◐ | ◐ | ○ | ○ | ◐ | ◐ | ○ | ◐ | ◐ | ○ |

## Leitura
- Ninguém tem ● no domínio 1 completo do brief: **ICP-Brasil, WhatsApp e LDAP direto** faltam em quase todos.
- Domínio 11 (identidade descentralizada) é ○ em quase todos; Keycloak (OID4VCI/VP) e Entra são exceções parciais.
- Domínio 8 (privacidade/conformidade) raramente é ●: ferramentas de direitos do titular e retenção são fracas no mercado.
- Os critérios de ● variam entre agentes; use a matriz para padrões, não para ranking fino.
