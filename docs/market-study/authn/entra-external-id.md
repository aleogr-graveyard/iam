# Microsoft Entra External ID

| Campo | Valor |
|---|---|
| Categoria | IDaaS / CIAM (configuração "external tenant" do Microsoft Entra), sucessor do Azure AD B2C; a mesma marca cobre B2B collaboration (convidados) em tenants de workforce |
| Site / repositório | https://learn.microsoft.com/en-us/entra/external-id/ · SDKs MSAL em https://github.com/AzureAD (serviço fechado) |
| Licença e modelo de negócio | Serviço proprietário em Azure. Cobrança por MAU: grátis até 50.000 MAU, depois preço por MAU; add-ons sem faixa gratuita: M2M Authentication (por transação), SMS Phone Authentication (por evento), Go-Local (residência de dados, por MAU; hoje só Austrália e Japão), ID Governance e GSA para convidados (workforce). Azure AD B2C: fim de venda a novos clientes em 01/05/2025 |
| Stack | Fechada (infraestrutura Entra ID/Azure). SDKs MSAL (.NET, JS, Java, Python, iOS, Android) e SDKs de "native authentication" |
| Maturidade | Entra ID (ex-Azure AD) é um dos maiores IdPs do mundo; External ID para clientes ficou GA em 2024; documentação ativa (páginas de recursos atualizadas em mar–jun/2026). Ainda há lacunas frente ao B2C (sem custom policies) |
| Acesso | 2026-11-03 |

> Limitação: `learn.microsoft.com` não abria pelo WebFetch, mas foi lido via `curl` (páginas oficiais citadas abaixo, com `ms.date`).

## Resumo
- CIAM gerenciado sobre a mesma plataforma do Entra ID: tenant "external" separado, user flows, branding, domínio personalizado, extensões HTTP em eventos.
- Atrativo: 50.000 MAU grátis e integração nativa com o ecossistema Microsoft/Azure.
- Faltam peças comuns em CIAM maduro: ID Protection e ID Governance indisponíveis em tenants externos, Conditional Access reduzido, sem custom policies do B2C, MFA limitado a e-mail OTP, SMS e passkey.
- Clientes B2C enfrentam migração forçada de produto (fim de venda em 2025) — sinal de risco de plataforma.
- Histórico de falhas graves de nuvem (actor tokens CVSS 10, nOAuth, Storm-0558).

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ◐ | E-mail+senha, e-mail OTP, username/alias (preview), Apple, Facebook, Google, conta Microsoft, federação Entra ID, OIDC genérico e SAML/WS-Fed. Federação só no modo navegador; native auth só contas locais. Passkey aparece como MFA. Sem LDAP/AD direto para clientes; sem ICP-Brasil |
| 2 | MFA e step-up | ◐ | E-mail OTP, SMS (add-on pago), passkey (FIDO2). Sem TOTP/push para clientes na tabela oficial. Conditional Access só com condições de plataforma e localização; sem risco (ID Protection indisponível). Step-up por "authentication context" |
| 3 | Ciclo de vida da conta | ◐ | Cadastro self-service com atributos customizados, verificação de e-mail, SSPR por e-mail OTP ou SMS, convites (preview, só administrativos), migração JIT de senhas (preview). SCIM para clientes e verificação de idade: não encontrado |
| 4 | Sessões e tokens | ◐ | Sign-in frequency, sessão persistente, front-channel logout, revogação via Graph. Back-channel logout, DPoP e rotação configurável de refresh token: não encontrados |
| 5 | Protocolos | ◐ | OIDC, OAuth 2.0 (code+PKCE, client credentials, device authorization, on-behalf-of, implicit), SAML (apps). Sem ROPC (substituído por native auth). CIBA, PAR, FAPI: não encontrados |
| 6 | Máquinas e agentes | ◐ | Client credentials como add-on cobrado por transação; credenciais federadas (workload identity federation) e certificados. Autenticação de agentes de IA específica: não encontrada no External ID |
| 7 | Detecção e defesa | ◐ | Proteção básica da plataforma; ID Protection "Not available" em tenants externos. Integração de proteção contra fraude (parceiros) citada na doc, não verificada |
| 8 | Privacidade e conformidade | ◐ | Localização do tenant na criação; Go-Local (Austrália, Japão). Certificações do Azure. Sem self-hosting; sem região Brasil dedicada no Go-Local |
| 9 | Experiência e administração | ◐ | Company branding, idiomas, custom URL domain, native authentication (UI própria no app), termos de uso no cadastro; painel Entra e Microsoft Graph. Personalização de UI mais limitada que a do B2C |
| 10 | Operação e migração | ◐ | Migração de usuários via Graph e JIT password migration (preview, com extensão customizada). Backup/export de hashes: não encontrado |
| 11 | Identidade descentralizada | ◐ | Produto separado Microsoft Entra Verified ID (conhecimento prévio); integração direta com external tenants não verificada |

## Diferenciais / exclusivos
- **50.000 MAU gratuitos** — maior faixa gratuita do grupo (Auth0 25.000; Cognito Lite 10.000 em conhecimento prévio).
- **Native authentication**: APIs/SDKs para telas 100% nativas no mobile/desktop/SPA sem navegador (troca segurança do navegador por controle de UX; doc lista as contrapartidas).
- **Custom authentication extensions**: eventos `OnAttributeCollectionStart`, `OnAttributeCollectionSubmit`, `OnOtpSend` (provedor de e-mail próprio) e `TokenIssuanceStart` (claims externos) chamando uma API REST.
- **Go-Local**: add-on de residência de dados por país.
- Mesma plataforma do workforce: app roles, grupos, RBAC no token, Graph API unificada.

## O que a comunidade elogia (com evidência)
- Preço de entrada (50k MAU grátis) — página oficial de preços.
- Integração com o ecossistema Microsoft (MSAL, Azure) — guias de migração e posts sobre B2C → External ID (fontes secundárias).
- Pouca discussão pública: busca no HN por "entra external id" retornou só posts com ≤ 2 pontos (sinal de baixa tração em comunidades de devs).

## Do que a comunidade se queixa (com evidência)
- **Migração forçada do B2C** e incerteza de datas: fim de venda em 01/05/2025; blogs de terceiros falam em fim do B2C P2 em 15/03/2026; resposta no Microsoft Q&A diz que não há planos de aposentar o B2C antes de 2030 — informação conflitante.
- **Paridade incompleta com o B2C**: sem Identity Experience Framework/custom policies (conhecimento prévio), Conditional Access e governança reduzidos — tabela oficial "Not available".
- **Doc confusa** entre workforce, B2B, B2C e external tenants (vários artigos de terceiros só para explicar as diferenças).
- **Lock-in**: autoridade `<tenant>.ciamlogin.com`, MSAL e Graph específicos.

## Histórico de segurança
| ID / caso | Ano | Classe | Lição para nós |
|---|---|---|---|
| nOAuth (Descope) | 2023 | Claim `email` mutável e não verificado usado como identificador por apps multi-tenant → tomada de conta | Identificar usuários federados por (`iss`, `sub`), nunca por e-mail; vincular por e-mail só com provedores que verificam (já é [BOA PRÁTICA] do brief) |
| Storm-0558 (conhecimento prévio) | 2023 | Chave de assinatura de contas de consumidor roubada e aceita para tokens corporativos por falha de validação de escopo da chave | Chaves por tenant/finalidade; validar emissor e escopo da chave; HSM/KMS e rotação; logs de emissão disponíveis ao cliente |
| CVE-2025-55241 (actor tokens) | 2025 | Tokens internos não documentados + validação de tenant falha na Azure AD Graph legada: um token do tenant do atacante permitia agir como qualquer usuário (inclusive admin global) de outro tenant; sem logs no tenant vítima, sem Conditional Access, não revogáveis por 24 h. CVSS 10 | Toda API valida o tenant do token contra o recurso; nenhum token "interno" fora das políticas e da auditoria; desligar APIs legadas; testes automáticos de isolamento entre tenants |
| CVE-2025-59218, CVE-2025-59246, CVE-2026-24305, CVE-2026-26148 | 2025–26 | Elevação de privilégio no Entra ID (CVSS 8,1–9,8; detalhes não publicados) | Falhas críticas recorrentes em serviços de identidade de nuvem; publicar advisories detalhados é diferencial de transparência |
| Midnight Blizzard (conhecimento prévio) | 2024 | Tenant de teste legado sem MFA + abuso de app OAuth com permissões elevadas | Inventário de tenants/apps; alertas de configuração fraca (premissa 4) inclusive para tenants de teste |

## Arquitetura e decisões notáveis
- External ID é uma **configuração de tenant** do Entra (external vs workforce), não um produto separado como o B2C; apps em tenant externo são sempre single-tenant.
- Personalização por **user flows** + **extensões HTTP** em pontos fixos, em vez de motor de políticas XML (B2C).
- Native auth: API própria de fluxo (não OAuth padrão de redirecionamento); a doc registra trade-offs (menos proteção anti-phishing do navegador, sem SSO).
- Permissões do Graph restritas em tenants externos (só `offline_access`, `openid`, `User.Read` + APIs próprias).
- Cobrança agregada por assinatura Azure, somando MAU de todos os tenants ligados.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- **Multi-tenancy**: um tenant externo por cliente/ambiente; para B2B dentro de um app, documentação não descreve "organizações" de primeira classe (não encontrado); recomenda-se grupos/app roles ou múltiplos tenants.
- **Extensibilidade**: custom authentication extensions (REST), Microsoft Graph, Azure Functions/Logic Apps (conhecimento prévio).
- **Configuração**: portal Entra, Graph API, Terraform/Bicep (conhecimento prévio).
- **Bancos**: não se aplica (SaaS).
- **Soberania**: região do tenant definida na criação; Go-Local só Austrália e Japão; sem opção on-premises.

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **não** no External ID (ID Governance indisponível em tenants externos). No workforce, Entitlement Management/PIM têm aprovação em múltiplos estágios (conhecimento prévio), sem quórum M de N documentado.
- **Autorização sensível ao estado / transições declarativas**: **não**. Autorização limitada a app roles, grupos e claims no token (RBAC/ABAC na aplicação).

## Implicações para o projeto
1. **Adotar**: pontos de extensão HTTP bem definidos por evento (pré-coleta de atributos, pós-submit, envio de OTP, emissão de token) — simples e suficiente para a maioria; combina com a premissa 18.
2. **Adotar**: "provedor de envio" substituível (e-mail/SMS/WhatsApp) como evento configurável (`OnOtpSend`).
3. **Oportunidade**: entregar sem add-on o que a Microsoft retira dos tenants externos — detecção de risco, MFA adaptativo, TOTP, governança.
4. **Oportunidade**: residência de dados por país/região como configuração (Brasil incluso), não add-on restrito.
5. **Evitar**: identificar contas federadas por e-mail (nOAuth); usar par `iss`+`sub`.
6. **Evitar**: tokens ou credenciais internas que escapam da política, da revogação e da auditoria (actor tokens); teste de isolamento entre tenants como suíte obrigatória de CI.
7. **Avaliar**: native authentication (API de fluxo sem navegador) como modo opcional para mobile/desktop (premissa 15), com alertas sobre os trade-offs.
8. **Evitar**: rupturas de produto que forçam migração de clientes; versionamento e ciclo de vida previsíveis (seção 4.2).

## Fontes
- https://learn.microsoft.com/en-us/entra/external-id/customers/concept-supported-features-customers — 2026-11-03 — oficial (ms.date 2026-03-30)
- https://learn.microsoft.com/en-us/entra/external-id/customers/concept-authentication-methods-customers — 2026-11-03 — oficial (ms.date 2026-03-27)
- https://learn.microsoft.com/en-us/entra/external-id/customers/concept-native-authentication — 2026-11-03 — oficial (ms.date 2026-06-22)
- https://learn.microsoft.com/en-us/entra/external-id/customers/how-to-custom-url-domain — 2026-11-03 — oficial
- https://learn.microsoft.com/en-us/entra/external-id/customers/whats-new-docs — 2026-11-03 — oficial (JIT password migration preview)
- https://learn.microsoft.com/en-us/entra/external-id/external-identities-pricing — 2026-11-03 — oficial (ms.date 2026-06-22; add-ons)
- https://azure.microsoft.com/en-us/pricing/details/microsoft-entra-external-id/ — 2026-11-03 — oficial (0–50.000 MAU grátis)
- https://www.schneider.im/microsoft-azure-ad-b2c-end-of-sale-to-new-customers/ — 2026-11-03 — blog de parceiro
- https://learn.microsoft.com/en-us/answers/questions/2265030/will-i-be-affected-by-azure-ad-b2c-end-of-sale-if — 2026-11-03 — fórum (não oficial)
- https://guptadeepak.com/ciam-compass/changelog/azure-ad-b2c-eol-march-2026/ — 2026-11-03 — blog de terceiro
- https://skycloak.io/blog/azure-ad-b2c-end-of-sale-support-timeline/ — 2026-11-03 — concorrente (Keycloak hospedado)
- https://www.securityweek.com/researchers-flag-account-takeover-flaw-in-microsoft-azure-ad-oauth-apps/ — 2026-11-03 — imprensa (nOAuth)
- https://thehackernews.com/2023/06/critical-noauth-flaw-in-microsoft-azure.html — 2026-11-03 — imprensa
- https://www.thestack.technology/critical-azure-vulnerability-cve-2025-55241-entraid/ — 2026-11-03 — imprensa
- https://www.itpro.com/security/microsoft-entra-id-vulnerability — 2026-11-03 — imprensa
- https://services.nvd.nist.gov/rest/json/cves/2.0?keywordSearch=entra%20id — 2026-11-03 — NVD
- https://hn.algolia.com/api/v1/search?query=entra%20external%20id — 2026-11-03 — comunidade (baixa tração)
