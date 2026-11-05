# Descope

| Campo | Valor |
|---|---|
| Categoria | CIAM/IDaaS "no/low-code" com orquestração visual de jornadas (**Flows**), B2B multi-tenant, federação (IdP OIDC/SAML/WS-Fed) e "Agentic Identity Hub" |
| Site / repositório | https://www.descope.com · docs https://docs.descope.com · SDKs em https://github.com/descope (serviço fechado) |
| Licença e modelo de negócio | Serviço proprietário; SDKs abertos (licença não verificada). **Free** (7.500 MAU, 10 tenants ativos, 3 conexões SSO, 10 mil trocas M2M), **Pro** US$ 249/mês (10 mil MAU, 35 tenants, 5 SSO, *SSO Setup Suite*), **Growth** US$ 799/mês (25 mil MAU, 100 tenants, 10 SSO, SCIM, FGA, migração de SSO, **multi-região**), Enterprise. Extras: US$ 0,05/MAU, US$ 1/tenant ativo, US$ 50/conexão SSO |
| Stack | Fechada; AWS (docs de multi-região). Linguagem do backend: não verificada. "Descope Engine" roda dentro da rede do cliente para conectores |
| Maturidade | Seed de US$ 53 mi em fev/2023 (HN 34851829). `@descope/node-sdk` ~1,3 mi e `@descope/web-component` ~1,1 mi downloads/mês (npm, out/2026). G2 4,8 com ~83 avaliações (via busca) |
| Acesso | 2026-11-03 |

> Limitações: GitHub e Reddit bloqueados nesta sessão; evidência de comunidade escassa (HN quase sem discussões).

## Resumo
- Diferencial central: **Flows**, editor visual (arrastar e soltar) de telas, ações, condições e conectores; o fluxo é a orquestração e termina emitindo o JWT de sessão.
- B2B nativo: **projeto** (isolamento total) → usuários no projeto → **tenants** com SSO, SCIM, papéis, estilo e políticas próprias; **sub-tenants** hierárquicos.
- Opção rara: **Tenant User Isolation** (mesmo e-mail vira identidade separada por tenant dentro do mesmo projeto).
- Protocolos mais amplos do grupo: IdP OIDC/SAML/WS-Fed, CIBA, PAR, DPoP, device code, token exchange, XAA/ID-JAG para agentes.
- Residência de dados em 7 regiões (inclui **SA, São Paulo**) e nuvem privada (FedRAMP).

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ● | OTP (e-mail/SMS/voz), link mágico, *enchanted link*, *embedded link*, senha, passkeys, social, SSO SAML/OIDC por tenant (vários IdPs por tenant), **nOTP via WhatsApp** (login com um clique), perguntas de segurança, push. LDAP direto e ICP-Brasil: não encontrados |
| 2 | MFA e step-up | ● | TOTP (auth apps), SMS/e-mail OTP, WhatsApp, passkeys, push, códigos de recuperação; **MFA adaptativo** por dispositivo confiável, reputação de IP e viagem impossível montado em Flows; step-up documentado |
| 3 | Ciclo de vida da conta | ● | Cadastro, verificação, recuperação, convites, JIT no SSO, SCIM por tenant, merge de identidades SSO (com página de risco), export de usuários e tenants, impersonação. Verificação de idade: só via conectores de terceiros (não verificado) |
| 4 | Sessões e tokens | ● | Session token + refresh token (JWT), validação offline, introspecção, **criptografia de JWT**, **DPoP (RFC 9449)** em Inbound Apps e clientes agênticos, política de sessão por tenant. Back-channel logout: não verificado |
| 5 | Protocolos | ● | IdP OIDC e SAML e **WS-Fed**, SP SAML/OIDC, SCIM, authorization code, client credentials, JWT bearer, refresh, **CIBA**, **device code**, **PAR**, **token exchange**, consentimento. FAPI: não encontrado |
| 6 | Máquinas e agentes | ● | Access keys/clientes, Inbound Apps, **Agentic Identity Hub** (registro de servidores MCP, clientes de workload, conexões de terceiros com tokens guardados, gateways), **XAA/ID-JAG** (IdP corporativo decide qual agente acessa qual MCP), *Agent Governance* com pedidos e aprovação |
| 7 | Detecção e defesa | ◐ | Viagem impossível, reputação de IP, dispositivo confiável em Flows; bot/fraude sobretudo via **conectores** de terceiros. Rate limiting nativo: não verificado |
| 8 | Privacidade e conformidade | ● | Região por projeto (US, EU, AP, CA, UK, SG, **SA**) fixa após criação; nuvem privada; FedRAMP; mascaramento de PII em audit logs; GDPR |
| 9 | Experiência e administração | ● | Flows + biblioteca de fluxos, estilos por tenant, **widgets** embutíveis (usuários, papéis, SSO, perfil do tenant, audit), Admin Portal, audit trail com streaming, i18n |
| 10 | Operação e migração | ◐ | Versionamento de flows e de projeto, **Terraform/Pulumi**, ambientes em GitHub/GitLab, guias de migração (Auth0, Cognito, Azure AD B2C, Firebase, Keycloak, Clerk, Stytch, Ping), migração de sessões e de SSO "sem downtime". Só SaaS (ou nuvem privada gerida) |
| 11 | Identidade descentralizada | ? | Não encontrado na doc; possível via conectores (não verificado) |

## Diferenciais / exclusivos
- **Flows visuais** com quatro blocos (telas, ações, condições, conectores), **interativos** e **de gestão** (automação de backend), versionados com notas e restauração.
- **Política do tenant só pode ser mais restritiva** que a do projeto; usuário em vários tenants recebe a combinação mais restritiva de políticas de senha.
- **Tenant User Isolation**: identidade separada por tenant sem criar outro projeto; exige passar o tenant na autenticação.
- **Sub-tenants** com herança configurável de usuários/papéis (só para baixo) e papéis herdados de ancestrais (ativação via suporte).
- **nOTP** via WhatsApp sem custo de SMS.
- **Descope Engine** dentro da rede do cliente para conectores sem abrir portas.
- **XAA/ID-JAG** e **Agent Governance** (pedido de acesso de agente aprovado por humano).

## O que a comunidade elogia (com evidência)
- Flows facilitam configurar jornadas e permitem muita personalização (avaliações AWS Marketplace, via busca; G2 4,8). Evidência de plataformas de avaliação com possível incentivo.
- Pouco a criticar segundo avaliadores do G2 (via busca).

## Do que a comunidade se queixa (com evidência)
- Funcionalidades fora de Flows ficam atrás: Google One Tap fora de Flows não foi concluído (avaliação AWS Marketplace, via busca).
- Discussão independente escassa: HN quase sem menções; Reddit inacessível. **Não encontrei** queixas consolidadas.
- Risco estrutural (inferência, não evidência): lógica de autenticação vive no console do fornecedor; mitigado por Terraform e versionamento, mas aumenta o lock-in.

## Histórico de segurança
- CVEs/advisories: **não encontrados** no OSV para `@descope/node-sdk`, `@descope/web-component`, `@descope/nextjs-sdk`, `@descope/react-sdk`, PyPI `descope`, Go `github.com/descope/go-sdk`; busca web sem resultados.
- Lição (de design, não de incidente): a doc dedica página ao **risco de mesclar identidades SSO** e exige que políticas de tenant só endureçam — confirmar ambos como regras de modelo de ameaça.

## Arquitetura e decisões notáveis
- Fluxo como única orquestração: entrada (ID do flow ou federação OIDC/SAML) → saída (JWT de sessão).
- Conectores síncronos ou assíncronos, resultado no contexto do flow para condições seguintes.
- Configuração como código (Terraform v1 com recursos separados), mas dados dinâmicos (usuários, tenants, conexões SSO) fora do Terraform.
- Projeto preso a uma região; multi-região = vários projetos + domínios de autenticação por região + roteamento na aplicação.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- **Níveis**: projeto (isolamento total, região) → tenants (clientes) → sub-tenants (unidades, franquias, revenda).
- Usuário no projeto; membership em 0..N tenants; **papéis por tenant** no JWT; papéis de projeto + papéis de tenant; papel *Tenant Admin* delega SSO, usuários e impersonação.
- Por tenant: métodos de autenticação (só mais restritivos), política de senha/OTP/link mágico, SSO (vários IdPs), SCIM, estilos, sessão, atributos customizados, domínios.
- Extensão: conectores (dezenas), scriptlets, management flows, webhooks/streaming de auditoria.
- Bancos: não se aplica. Soberania: 7 regiões + nuvem privada; dados não migram entre regiões.

## Verificação de diferenciais candidatos do projeto
- Aprovação multipartes (M de N): **não**. Há aprovação individual em *Agent Governance* (pedido de cliente para agente) e aprovação do usuário via CIBA; sem quórum.
- Autorização sensível ao estado / transições: **não** encontrada (RBAC/ReBAC/ABAC sem estados; avaliado em AuthZ).

## Implicações para o projeto
1. **Adotar** a regra "tenant/organização só endurece a política do nível acima" e "várias políticas aplicáveis → a mais restritiva"; casa com a premissa de composabilidade e com alertas.
2. **Adotar** isolamento de identidade por tenant como opção de configuração (equivalente a "promover organização a tenant" sem migração).
3. **Oportunidade**: jornadas declarativas e versionadas como **código** (arquivo revisável e testável) em vez de editor só visual; editor visual como camada sobre o mesmo artefato.
4. **Adotar** sub-organizações hierárquicas com herança explícita e unidirecional.
5. **Adotar** região fixa por tenant com roteamento por domínio; considerar região Brasil desde o início.
6. **Adotar** protocolos amplos (CIBA, PAR, DPoP, token exchange, ID-JAG) como referência de escopo para domínios 5 e 6.
7. **Evitar** recursos que só existem dentro do orquestrador visual (queixa do One Tap).

## Fontes
- https://docs.descope.com/llms.txt — 2026-11-03 — oficial
- https://docs.descope.com/b2b.md, /management/tenant-management.md, /management/tenant-management/tenant.md, /management/tenant-management/sub-tenants.md — 2026-11-03 — oficial
- https://docs.descope.com/management/project-settings/multi-regional.md, /how-to-deploy-to-production/multi-region-architecture.md, /how-to-deploy-to-production/private-cloud.md — 2026-11-03 — oficial
- https://docs.descope.com/flows.md, /flows/intro-to-flows/flow-versioning.md, /managing-environments/terraform.md, /connectors.md — 2026-11-03 — oficial
- https://docs.descope.com/mfa-and-step-up/mfa/adaptive-mfa.md, /auth-methods/push.md, /auth-methods/notp.md, /auth-methods/embedded-link.md, /sessions.md — 2026-11-03 — oficial
- https://docs.descope.com/identity-federation/inbound-apps/dpop.md, /identity-federation/inbound-apps/grant-types/ciba.md, /agentic-identity-hub/agent-governance.md, /agentic-identity-hub/enterprise-managed-authorization.md — 2026-11-03 — oficial
- https://www.descope.com/pricing — 2026-11-03 — oficial
- https://api.osv.dev/v1/query (pacotes listados acima) — 2026-11-03 — base pública
- https://www.g2.com/products/descope e avaliações AWS Marketplace (via WebSearch) — 2026-11-03 — sites de avaliação, possível viés de incentivo
- https://news.ycombinator.com/item?id=34851829 — 2026-11-03 — comunidade
- https://api.npmjs.org/downloads/point/last-month/@descope/node-sdk — 2026-11-03 — npm
