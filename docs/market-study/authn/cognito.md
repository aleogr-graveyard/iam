# Amazon Cognito

| Campo | Valor |
|---|---|
| Categoria | IDaaS / CIAM gerenciado da AWS: *user pools* (diretório + OIDC/OAuth) e *identity pools* (troca de identidade por credenciais AWS temporárias) |
| Site / repositório | https://aws.amazon.com/cognito/ · docs https://docs.aws.amazon.com/cognito/ · SDKs AWS e Amplify (serviço fechado) |
| Licença e modelo de negócio | Serviço proprietário. Por MAU em três planos de recursos (nov/2024): **Lite** (legado, preço em faixas), **Essentials** (padrão para novos pools; ~US$ 0,015/MAU segundo fontes secundárias) e **Plus** (~US$ 0,02/MAU, fontes secundárias). M2M cobrado à parte (conhecimento prévio). Replicação multi-região é add-on por MAU por região réplica (jun/2026). SMS cobrado via SNS |
| Stack | Fechada. Extensão por AWS Lambda; e-mail via SES; SMS via SNS/End User Messaging; proteção por AWS WAF |
| Maturidade | Lançado em 2014 (conhecimento prévio); uso amplo em workloads AWS; percepção de produto "semiabandonado" até 2024, com retomada (planos novos, passkeys, managed login em nov/2024; multi-região em jun/2026) |
| Acesso | 2026-11-03 |

> Limitação: `docs.aws.amazon.com` e `aws.amazon.com` não abriam pelo WebFetch; usei buscas (trechos das páginas oficiais), NVD e HN. Itens "(conhecimento prévio)" não foram reconferidos.

## Resumo
- Autenticação barata e integrada à AWS (API Gateway, ALB, AppSync, IAM via identity pools).
- Extensão por gatilhos Lambda em quase todos os pontos do fluxo, inclusive migração preguiçosa de usuários.
- Desde nov/2024: managed login (UI hospedada nova), passkeys, login sem senha por e-mail/SMS OTP (Essentials) e proteção contra ameaças (Plus).
- Queixas fortes: documentação, configurações imutáveis após criar o pool, impossibilidade de exportar hashes de senha, ausência de backup nativo, SAML só como SP, sem SCIM.
- Segurança: a maioria dos problemas públicos vem de **padrões permissivos** (atributos graváveis pelo próprio usuário) e de configuração de IAM em torno do serviço.

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ◐ | Senha (SRP ou direta), passkeys e OTP por e-mail/SMS como primeiro fator (Essentials+), social (Google, Facebook, Apple, Amazon), SAML e OIDC de entrada. Link mágico nativo: não (só via custom auth challenge). LDAP/AD direto: não (via SAML/ADFS). ICP-Brasil: não |
| 2 | MFA e step-up | ◐ | TOTP, SMS, e-mail (Essentials+), WebAuthn/passkey. Push: não. Adaptativo por risco só no Plus (threat protection). Step-up nativo: não encontrado (implementar via custom auth/Lambda) |
| 3 | Ciclo de vida da conta | ◐ | Cadastro, verificação de e-mail/telefone, recuperação, criação por admin com convite, migração JIT (gatilho *user migration*), importação CSV (sem senha). Vinculação de contas federadas via API `AdminLinkProviderForUser`. SCIM: não. Verificação de idade: não |
| 4 | Sessões e tokens | ◐ | Revogação de refresh token e global sign-out; rotação de refresh token adicionada em 2025 (conhecimento prévio); personalização de access token (Essentials+). Back-channel logout e DPoP: não encontrados |
| 5 | Protocolos | ◐ | OAuth 2.0/OIDC (code+PKCE, implicit, client credentials), SAML só como SP (não atua como IdP SAML). Device flow, CIBA, PAR, FAPI, SCIM: não encontrados |
| 6 | Máquinas e agentes | ◐ | Client credentials com escopos de resource server (M2M cobrado). Identity pools emitem credenciais AWS temporárias (identidade de workload dentro da AWS). Agentes de IA: não encontrado |
| 7 | Detecção e defesa | ◐ | Plus: detecção de credenciais comprometidas, autenticação adaptativa por risco, export de eventos. Integração com AWS WAF; quotas de API. Bloqueio por tentativas é fixo, pouco configurável (conhecimento prévio) |
| 8 | Privacidade e conformidade | ◐ | Região escolhida por pool; GovCloud; conformidades da AWS. Export completo de dados e hashes: não |
| 9 | Experiência e administração | ◐ | Managed login com editor visual (Essentials+), Hosted UI clássica (Lite), console AWS, i18n limitada (conhecimento prévio). Webhooks/eventos: via Lambda/CloudTrail, não webhooks nativos |
| 10 | Operação e migração | ◐ | Migração de entrada via Lambda. Sem backup/restore nativo de pool; sem export de hashes; replicação multi-região (add-on, jun/2026, só infraestrutura "next-gen" e algumas regiões) |
| 11 | Identidade descentralizada | ○ | Não encontrado |

## Diferenciais / exclusivos
- **Identity pools**: troca de token (próprio, social, SAML, OIDC) por credenciais AWS temporárias com papéis IAM — acesso direto a S3/DynamoDB sem backend.
- **Gatilhos Lambda abrangentes**: pre sign-up, pre/post authentication, pre token generation (com versões para personalizar access token), define/create/verify auth challenge (fluxos customizados), user migration, custom message, custom email/SMS sender.
- **Preço por MAU baixo** e cobrança unificada na conta AWS.
- **Replicação multi-região** gerenciada (2026) para usuários, máquinas e configuração.

## O que a comunidade elogia (com evidência)
- Custo: "One big pro about cognito.. can't beat the price" — HN 49478091.
- Integração com serviços AWS (API Gateway authorizers, ALB, AppSync) — docs e posts de usuários (conhecimento prévio).

## Do que a comunidade se queixa (com evidência)
- **Sem export de hashes de senha** → migração de saída força reset para todos; workaround é migração JIT verificando a senha contra o Cognito — HN 49478091; https://securityboulevard.com/2022/02/how-to-migrate-from-amazon-cognito/ (2022); guia de 2026 de terceiro.
- **Configurações imutáveis**: mudar sensibilidade a maiúsculas do e-mail exige criar outro pool ("no way to migrate") — HN 49478091.
- **Documentação**: "three separate manuals, threw them in a blender" — HN 49478091 (post de ago/2026, 193 pontos, 132 comentários; um comentário diz que o texto parece gerado por IA).
- **Produto estagnado / sem backup**: "felt semi-abandoned", "Don't even get me started on backups" — HN 49478091.
- **Incidentes de disponibilidade**: "AWS Cognito is having issues and health dashboards are still green" (nov/2020, 492 pontos) — https://news.ycombinator.com/item?id=25209773; problemas em us-east-1 em out/2023.
- **SAML/SCIM limitados** para B2B (fonte concorrente, SSOJet).

## Histórico de segurança
CVEs do serviço em si: **não encontrados** no NVD (o serviço é gerenciado; a AWS publica boletins próprios). Casos relevantes:

| ID / caso | Ano | Classe | Lição para nós |
|---|---|---|---|
| Doyensec — atributos não restritos | 2023 | Usuário altera os próprios atributos (inclusive customizados) com o access token por padrão; apps que confiam nesses atributos sofrem tomada de conta/escalada | Atributos usados em decisão de autorização ou de identidade devem ser somente-leitura para o usuário **por padrão**; permissões de escrita por cliente explicitamente declaradas |
| Cobalt — troca de e-mail | 2023–24 | E-mail alterado sem verificação + normalização de maiúsculas na app → tomada de conta | Mudança de e-mail só vale após verificação (Cognito tem `RequireAttributesVerifiedBeforeUpdate`, mas opcional); normalizar identificadores de forma única e definida |
| Truesec — atributos customizados | 2023 | Escalada de privilégio por sobrescrita de atributo "admin" | Papel/privilégio nunca em atributo editável pelo titular |
| CVE-2024-28056 | 2024 | Amplify CLI: ao remover autenticação, a trust policy do papel IAM ficava `Allow` sem condição → `AssumeRoleWithWebIdentity` por qualquer um (CVSS 9,8) | Remoção de componente de segurança deve "falhar fechado"; ferramentas de provisionamento são parte da superfície |
| CVE-2024-45037 | 2024 | AWS CDK: `CognitoUserPoolAuthorizer` ignorava escopos → acesso mais amplo que o pretendido | Testes de autorização negativa (escopo ausente = negar) nos SDKs/integrações |
| Identity pools com acesso não autenticado (prática de pentest) | contínuo | Papéis IAM amplos para identidades "guest" | Alertar configurações perigosas (premissa 4) |

## Arquitetura e decisões notáveis
- Separação user pool (autenticação) × identity pool (autorização para recursos AWS).
- SRP (Secure Remote Password) como fluxo de senha padrão nos SDKs — senha não trafega (conhecimento prévio).
- Esquema de atributos definido na criação; vários parâmetros (aliases de login, case sensitivity, atributos obrigatórios) são imutáveis.
- Planos de recursos por pool (Lite/Essentials/Plus) — recursos de segurança vinculados ao plano.
- Extensão síncrona por Lambda com timeout curto (5 s, conhecimento prévio) no caminho crítico do login.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- **Multi-tenancy**: a AWS documenta 4 padrões — pool por tenant, app client por tenant, grupo por tenant, atributo customizado por tenant. Recomenda atributo de tenant imutável, mapeamento 1:1 IdP externo ↔ app client e validação de tenancy com Amazon Verified Permissions. Admite que atributo customizado serve para isolamento "superficial"; MFA e branding são por pool ou app client.
- **Extensibilidade**: gatilhos Lambda; eventos via CloudTrail.
- **Configuração**: console, API, CloudFormation/CDK/Terraform.
- **Bancos**: não se aplica (SaaS).
- **Soberania**: região escolhida por pool; replicação multi-região opcional em subconjunto de regiões; sem self-hosting.

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **não**. Nada no Cognito; na AWS, aprovações existem em outros serviços (ex.: IAM Identity Center/temporary elevated access) fora do escopo.
- **Autorização sensível ao estado / transições declarativas**: **não**. Cognito delega autorização a grupos/claims e ao Amazon Verified Permissions (Cedar), que avalia atributos de contexto mas não modela transições.

## Implicações para o projeto
1. **Adotar**: catálogo amplo de ganchos (incluindo *user migration*, *custom challenge*, *custom sender*) — habilita migração sem reset e provedores de envio substituíveis.
2. **Evitar**: configurações imutáveis após criação; toda opção relevante (normalização de identificador, aliases, atributos obrigatórios) deve ser alterável com migração assistida.
3. **Oportunidade**: export completo (usuários, hashes com parâmetros do algoritmo, MFA quando possível) e backup/restore de tenant — queixa central contra o Cognito.
4. **Adotar**: atributos com permissão de leitura/escrita por cliente e **somente leitura por padrão** para atributos sensíveis; verificação obrigatória antes de trocar e-mail/telefone.
5. **Adotar**: documentação dos padrões de multi-tenancy com trade-offs explícitos (isolamento por pool vs atributo), como faz a AWS, mas com isolamento forte nativo.
6. **Oportunidade**: oferecer SAML IdP, SCIM, device flow e back-channel logout — lacunas notórias do Cognito em B2B.
7. **Adotar**: alertas de configuração para combinações perigosas (identidades não autenticadas com permissões amplas, escopos ignorados).

## Fontes
- https://aws.amazon.com/about-aws/whats-new/2024/11/new-feature-tiers-essentials-plus-amazon-cognito — 2026-11-03 — oficial
- https://docs.aws.amazon.com/cognito/latest/developerguide/cognito-sign-in-feature-plans.html — 2026-11-03 — oficial (via trechos de busca)
- https://aws.amazon.com/cognito/pricing/ — 2026-11-03 — oficial (valores não extraídos)
- https://blog.logto.io/amazon-cognito-pricing — 2026-11-03 — concorrente (valores por MAU)
- https://aws.amazon.com/blogs/aws/improve-your-app-authentication-workflow-with-new-amazon-cognito-features/ — 2026-11-03 — oficial
- https://aws.amazon.com/es/about-aws/whats-new/2026/06/amazon-cognito-multi-region/ — 2026-11-03 — oficial
- https://infoq.com/news/2026/06/cognito-replication-aws — 2026-11-03 — imprensa
- https://aws.amazon.com/cognito/faqs/ — 2026-11-03 — oficial
- https://docs.aws.amazon.com/cognito/latest/developerguide/multi-tenancy-security-recommendations.html — 2026-11-03 — oficial
- https://docs.aws.amazon.com/cognito/latest/developerguide/custom-attribute-based-multi-tenancy.html — 2026-11-03 — oficial
- https://ssojet.com/blog/cognito-saml-enterprise-sso-limits — 2026-11-03 — concorrente
- https://blog.doyensec.com/2023/01/24/tampering-unrestricted-user-attributes-aws-cognito.html — 2026-11-03 — pesquisa de segurança
- https://www.cobalt.io/blog/full-account-takeover-via-aws-cognito-misconfiguration — 2026-11-03 — pesquisa de segurança
- https://truesec.com/hub/blog/aws-cognito-token-security-one-step-closer — 2026-11-03 — pesquisa de segurança
- https://services.nvd.nist.gov/rest/json/cves/2.0?cveId=CVE-2024-28056 e CVE-2024-45037 — 2026-11-03 — NVD
- https://news.ycombinator.com/item?id=49478091 — 2026-11-03 — comunidade
- https://news.ycombinator.com/item?id=25209773 — 2026-11-03 — comunidade
- https://securityboulevard.com/2022/02/how-to-migrate-from-amazon-cognito/ — 2026-11-03 — artigo de 2022, autor/viés não verificado
- https://guptadeepak.com/ciam-compass/guides/migrating-from-cognito/ — 2026-11-03 — blog de terceiro
- https://dev.to/raphael_jambalos/amazon-cognito-the-ugly-parts-and-our-workarounds-2k5p — 2026-11-03 — relato de usuário
