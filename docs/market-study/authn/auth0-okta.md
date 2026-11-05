# Auth0 / Okta (Customer Identity e Workforce Identity)

| Campo | Valor |
|---|---|
| Categoria | IDaaS. Auth0 = CIAM (clientes, B2C/B2B), também comercializado como "Okta Customer Identity"; Okta Workforce Identity = IAM corporativo (funcionários, SSO para SaaS, ciclo de vida, governança) |
| Site / repositório | https://auth0.com · https://www.okta.com · SDKs em https://github.com/auth0 e https://github.com/okta (serviço fechado) |
| Licença e modelo de negócio | Serviço proprietário (SaaS; Auth0 também em "Private Cloud" dedicada). Abertos: SDKs, bibliotecas (ex.: `node-jsonwebtoken`, `java-jwt`, `node-jws`) e o OpenFGA (doado à CNCF; base do Auth0 FGA). Auth0 cobra por MAU com planos Free, Essentials, Professional e Enterprise, separados em B2C e B2B; plano gratuito ampliado de 7.500 para 25.000 MAU em set/2024; reajuste forte nos planos pagos em nov/2023. Okta Workforce cobra por usuário/mês por produto (SSO, MFA, Lifecycle, Governance, Privileged Access etc.) |
| Stack | Fechada. Auth0: historicamente Node.js (extensibilidade em Node.js via Actions); detalhes de banco não publicados oficialmente (não verificado). Okta: Java (não verificado nesta sessão) |
| Maturidade | Auth0 fundada em 2013; Okta fundada em 2009 e listada na Nasdaq; Okta comprou a Auth0 por US$ 6,5 bi (anúncio mar/2021, HN 632 pontos). Líder de mercado em CIAM e workforce |
| Acesso | 2026-11-03 |

> Limitação desta pesquisa: `auth0.com` e `okta.com` não estavam liberados na rede da sessão para leitura direta; usei buscas na web (que retornam trechos das páginas oficiais), NVD (API) e Hacker News (API Algolia). Itens marcados "(conhecimento prévio)" não foram reconferidos nesta sessão.

## Resumo
- Referência de mercado em CIAM: Universal Login, Actions (Node.js serverless no pipeline), Organizations para B2B, Attack Protection e catálogo enorme de conexões.
- Okta Workforce cobre SSO corporativo, diretório, provisionamento (SCIM), MFA resistente a phishing (FastPass), governança e acesso privilegiado.
- Fortes em padrões (FAPI, PAR, JAR, RAR, mTLS, CIBA) e em identidade de agentes de IA (Auth0 for AI Agents GA nov/2025; Cross App Access/ID-JAG).
- Pontos fracos recorrentes: preço e "degraus" entre planos, recursos essenciais travados em planos caros, aprisionamento (export de hashes via suporte).
- Histórico de incidentes corporativos sérios (2022–2023), com lições diretas para suporte, sessões e divulgação.

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ◐ | Senha, passwordless (link mágico e código por e-mail/SMS), passkeys, dezenas de provedores sociais, SAML/OIDC/WS-Fed corporativo, AD/LDAP via conector on-premises. Sem ICP-Brasil nativo (só via certificado cliente/mTLS ou IdP externo) |
| 2 | MFA e step-up | ● | Push (Guardian/Okta Verify), TOTP, WebAuthn (roaming e plataforma), SMS, voz, e-mail, códigos de recuperação, Duo; MFA adaptativo por risco (add-on Enterprise); step-up via `acr_values`. Okta: FastPass ligado ao dispositivo. WhatsApp não encontrado como fator nativo |
| 3 | Ciclo de vida da conta | ● | Cadastro, verificação, reset, vinculação de contas, convites (Organizations), migração preguiçosa via Custom DB, SCIM de entrada para conexões corporativas; Okta Lifecycle Management com SCIM de saída. Verificação de idade: não encontrado |
| 4 | Sessões e tokens | ◐ | Rotação de refresh token com detecção de reuso, revogação, OIDC Back-Channel Logout, Universal Logout (Okta). DPoP em Early Access (docs "sender constraining") |
| 5 | Protocolos | ● | OAuth 2.0, OIDC, SAML, WS-Fed, SCIM, WebAuthn, device flow, CIBA, PAR, JAR, RAR, mTLS, `private_key_jwt`; certificado FAPI (perfil FAPI 1 Advanced) no add-on Highly Regulated Identity |
| 6 | Máquinas e agentes | ● | Client credentials (M2M cobrado à parte), Auth0 for AI Agents (Token Vault, CIBA assíncrono, FGA para RAG), Okta Cross App Access (ID-JAG, rascunho IETF) |
| 7 | Detecção e defesa | ● | Bot detection, breached password detection, brute-force e suspicious IP throttling (Auth0 Attack Protection); ThreatInsight e Identity Threat Protection (Okta). Parte travada por plano |
| 8 | Privacidade e conformidade | ◐ | Regiões (EUA, UE, AU, JP etc.), Private Cloud, certificações (SOC 2, ISO 27001, HIPAA BAA etc.) (conhecimento prévio). Sem self-hosting; soberania limitada às regiões oferecidas |
| 9 | Experiência e administração | ● | Universal Login com branding/templates, Forms, i18n, painel, Log Streams, Management API, Terraform provider, Deploy CLI |
| 10 | Operação e migração | ◐ | Importação em massa com hashes, migração preguiçosa (Custom DB). Exportar hashes exige ticket de suporte (relato no HN). Backup/restore pelo cliente: não oferecido (é SaaS) |
| 11 | Identidade descentralizada | ○ | Auth0 lançou "Verifiable Credentials" em 2022 (HN 229 pontos); situação atual não verificada |

## Diferenciais / exclusivos
- **Actions**: código Node.js versionado executado em gatilhos (login, pré/pós-registro, troca de senha, M2M); substituiu Rules/Hooks.
- **Organizations**: multi-tenancy B2B de primeira classe (login por organização, convites, papéis e conexões por organização), até 2.000.000 organizações por tenant mediante pedido.
- **Highly Regulated Identity**: FAPI certificado, PAR, JAR, RAR (Dynamic Linking para transações), mTLS — referência para Open Finance.
- **Auth0 for AI Agents** (GA 19/nov/2025): Token Vault (guarda e renova tokens OAuth de terceiros para o agente), autorização assíncrona por CIBA, FGA para RAG.
- **Okta Cross App Access (XAA)**: o agente troca o ID token por um ID-JAG e obtém acesso a outro app sem novo consentimento; política central do administrador.
- **Universal Logout** e sinais compartilhados (CAEP/Shared Signals) no Workforce.

## O que a comunidade elogia (com evidência)
- Experiência de desenvolvimento e documentação da Auth0 pré-aquisição: comentários no HN "Goodbye Auth0" ("Auth0 really did walk the walk in terms of developer experience") — https://news.ycombinator.com/item?id=39380790.
- Amplitude de integrações e padrões (FAPI, CIBA, RAR) — docs HRI.
- Permitir exportar hashes de senha (mesmo via ticket) é citado como raro e positivo frente ao Cognito — HN 49478091.

## Do que a comunidade se queixa (com evidência)
- **Preço**: aumento de até ~300% por MAU no B2C Essentials no fim de 2023 e "cliff" entre gratuito e pago (fontes concorrentes: SSOJet, Stytch, WorkOS — viés de concorrente). Anúncio oficial: https://auth0.com/blog/upcoming-pricing-changes-for-the-customer-identity-cloud/.
- **Recursos essenciais travados por plano**: RBAC, bot detection, vinculação de identidades fora do Free segundo WorkOS (concorrente); MFA adaptativo e HRI só no Enterprise.
- **"Paga-se muito pelo que é um JWT"**: comentário no HN 39380790 (opinião).
- **Disponibilidade**: "Auth0 has been down for almost 4 hours" (abr/2021, HN 195 pontos) — https://news.ycombinator.com/item?id=26880147.
- **Contagem de MAU por tenant** pode inflar a cobrança (guia de 2026, fonte secundária).
- **Pós-aquisição**: demissões e percepção de perda de cultura (post "Goodbye Auth0", HN 286 pontos).

## Histórico de segurança

### Incidentes notórios da Okta
| Quando | Incidente | Classe | Lição para nós |
|---|---|---|---|
| jan/2022 (divulgado mar/2022) | Lapsus$ acessou o laptop de um engenheiro de suporte terceirizado (Sitel/Sykes) por ~5 dias; até 366 clientes potencialmente afetados (2 de fato); Okta admitiu erro por demorar ~2 meses a divulgar | Terceiro com acesso privilegiado; atraso na divulgação | Acesso de suporte é superfície de ataque: mínimo privilégio, sessão curta, auditoria e alerta ao cliente. Ter política de divulgação com prazos definidos (processo, não código) |
| dez/2022 | Roubo de repositórios GitHub do Workforce Identity Cloud | Credenciais de repositório | Código não pode guardar segredos; assumir que o código-fonte vaza (para nós é aberto, então "segurança por obscuridade" está fora por definição) |
| ago/2023 | Engenharia social contra helpdesks de clientes para resetar MFA de super admins (onda Scattered Spider) (conhecimento prévio) | Recuperação mais fraca que o MFA | Reforça o [BOA PRÁTICA] do brief: recuperação/reset feito pelo suporte nunca mais fraco que o MFA; exigir verificação forte e dupla aprovação para reset de administradores |
| set–out/2023 | Sistema de suporte invadido com credencial de conta de serviço salva em conta Google pessoal; atacante leu arquivos HAR com tokens de sessão e sequestrou sessões de 5 clientes (1Password, BeyondTrust, Cloudflare); escopo ampliado depois para todos os usuários do suporte | Vazamento de tokens em artefatos de diagnóstico; detecção lenta (alerta em 29/set, confirmação ~2 semanas depois) | Sessões administrativas presas ao dispositivo/IP e curtas; sanitizar HAR/logs no recebimento; nunca registrar tokens; tratar alertas de clientes como incidente imediato |
| 2023–2024 | Resposta: "Project Bedrock" (90 dias só segurança) e "Secure Identity Commitment" (MDM obrigatório, timeout de sessões admin, restrição de acesso a casos de suporte) | — | Proteções do console admin (timeout, ligação ao dispositivo, step-up) como padrão, não opcional |
| jul–out/2024 | AD/LDAP Delegated Auth: chave de cache gerada com bcrypt sobre userId+username+senha; bcrypt trunca em 72 bytes, então com username ≥ 52 caracteres a senha não entrava no hash e um login em cache aceitava qualquer senha (sem MFA). Corrigido trocando para PBKDF2 | Uso incorreto de primitiva criptográfica | Nunca usar bcrypt como hash genérico/KDF de chave composta; testar limites de entrada; caches de credencial são caminho de autenticação e precisam do mesmo rigor |
| 2024–2026 | CVEs em agentes e clientes: Okta Verify Windows (CVE-2024-0980, CVE-2024-7061, CVE-2024-9191), Okta Verify iOS (CVE-2024-10327), Browser Plugin XSS (CVE-2024-0981), Access Gateway (dezenas em set–out/2026: injeção de comando, SQL/LDAP via atributos SAML, header de identidade sem verificação criptográfica — CVE-2026-78560) | Injeção, DLL hijacking, escalada local | Componentes on-premises/agents têm a maior densidade de falhas; atributos vindos de asserções SAML são entrada não confiável |

### Auth0 (serviço e SDKs)
| ID | Ano | Classe | Lição |
|---|---|---|---|
| CVE-2018-6873 | 2018 | Serviço não validava `aud` do JWT → escalada | Validação completa de claims (`aud`, `iss`, `exp`) obrigatória e testada |
| CVE-2018-6874 | 2018 | CSRF com flag Legacy Lock API | Modos legados precisam do mesmo nível de teste |
| CVE-2018-7307, CVE-2018-15121 | 2018 | SDKs sem `state` → CSRF no fluxo OAuth | SDKs próprios com `state`/PKCE/nonce obrigatórios |
| CVE-2025-46345 | 2025 | Account Link Extension não verificava assinatura do JWT | Vinculação de contas é fluxo crítico |
| CVE-2025-47275, CVE-2025-48951 | 2025 | Auth0-PHP: cookies de sessão fracos; desserialização insegura | SDKs também são superfície; auditar todos |
| CVE-2025-65945 | 2025 | `node-jws`: verificação de assinatura imprópria | Bibliotecas JOSE exigem testes de vetores negativos |
| CVE-2025-67490, CVE-2026-40155 | 2025–26 | nextjs-auth0: requisições simultâneas misturam sessões | Testes de concorrência em gestão de sessão |
| CVE-2026-85981..85983 | 2026 | AD/LDAP Connector: painel local sem auth, XSS, escalada local | Agentes on-premises com o mesmo rigor do núcleo |
| CVE-2022-23529 | 2022 | `jsonwebtoken` — relatada como RCE, depois rejeitada no NVD | Avaliar exploração real antes do alarde |

## Arquitetura e decisões notáveis
- Pipeline de login extensível por código do cliente (Actions) em sandbox; versões, segredos e dependências npm por Action.
- Universal Login hospedado (redirecionamento) recomendado sobre login embutido (Lock); domínio personalizado.
- Tenant = unidade de isolamento, configuração e região; ambientes (dev/stage/prod) são tenants separados.
- Private Cloud: implantação dedicada em AWS/Azure para clientes Enterprise.
- Okta: agentes on-premises (AD, LDAP, RADIUS, Access Gateway) fazem a ponte com sistemas legados; geram boa parte dos CVEs.
- Okta Identity Engine: políticas de autenticação por app com "assurance" (posse, conhecimento, inerência; resistência a phishing).

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- **Multi-tenancy**: tenants Auth0 (isolamento forte) + Organizations (B2B dentro do tenant; branding, conexões, papéis e MFA por organização). Limites de entidades por plano.
- **Extensibilidade**: Actions, Forms + Flows + Vault, Custom DB scripts, Custom Social Connections, Log Streams, Management API.
- **Configuração**: painel, Management API, Terraform, Deploy CLI (configuração como código).
- **Bancos**: não se aplica (SaaS; o cliente não escolhe).
- **Soberania**: escolha de região do tenant; Private Cloud; sem opção self-hosted — limitação para LGPD/soberania estrita.

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **parcial (só Workforce)**. Okta Identity Governance tem solicitações de acesso com sequência de aprovadores, mas o padrão é "todos aprovam" na sequência; para grupo, basta um dono. Para papéis de admin exige 2 aprovadores. Não há quórum M de N configurável documentado (thread no suporte pede "qualquer um de N" sem solução clara). Auth0: não encontrado.
- **Autorização sensível ao estado / transições declarativas**: **não**. Auth0 FGA (OpenFGA) faz ReBAC com condições (contexto), mas sem modelo de estados/transições. Workflows do Okta são automação, não política de transição.

## Implicações para o projeto
1. **Adotar**: pipeline de extensibilidade por eventos (pré/pós-registro, login, emissão de token, M2M) — mas por webhooks/WASM configuráveis, não JS do cliente no nosso processo.
2. **Adotar**: Organizations como conceito de primeira classe (nosso cliente multi-tenant da premissa 10), com MFA, conexões e branding por organização.
3. **Adotar**: perfil "regulado" (FAPI 2, PAR, JAR, RAR, mTLS, DPoP) como perfil de conformidade configurável — oportunidade de entregar sem add-on caro (Open Finance Brasil).
4. **Oportunidade**: não travar segurança básica (bot detection, senhas vazadas, RBAC, MFA adaptativo) em planos pagos; é a queixa nº 1.
5. **Oportunidade**: export de usuários com hashes pelo próprio painel/API, sem ticket — antiaprisionamento como argumento de venda.
6. **Evitar**: acesso de suporte/impersonação sem amarras — exigir aprovação, sessão curta, ligação ao dispositivo, auditoria visível ao tenant e redação automática de tokens em anexos.
7. **Evitar**: uso de bcrypt como KDF genérica e caches de credenciais "fallback"; incluir teste de limites de comprimento em todo hash.
8. **Adotar**: identidade de agentes (CIBA para aprovação assíncrona, Token Exchange, ID-JAG) no roadmap do domínio 6.

## Fontes
- https://auth0.com/blog/upcoming-pricing-changes-for-the-customer-identity-cloud/ — 2026-11-03 — oficial
- https://community.auth0.com/t/auth0-pricing-update-what-you-need-to-know/147913 — 2026-11-03 — oficial (free 25k MAU, set/2024)
- https://ssojet.com/blog/auth0-pricing-growth-penalty — 2026-11-03 — concorrente
- https://stytch.com/blog/auth0-2024-pricing-update/ — 2026-11-03 — concorrente
- https://workos.com/blog/auth0-pricing-how-it-works-and-compares-to-workos — 2026-11-03 — concorrente
- https://auth0.com/docs/secure/highly-regulated-identity — 2026-11-03 — oficial (via trechos de busca)
- https://auth0.com/docs/secure/sender-constraining — 2026-11-03 — oficial (DPoP Early Access)
- https://auth0.com/blog/auth0-for-ai-agents-generally-available/ — 2026-11-03 — oficial
- https://auth0.com/docs/support/policies/entity-limit-policy — 2026-11-03 — oficial
- https://developer.okta.com/docs/journeys/XAA-requesting-app/main/ — 2026-11-03 — oficial
- https://help.okta.com/en-us/Content/Topics/identity-governance/access-requests/rcar-conditions.htm — 2026-11-03 — oficial
- https://support.okta.com/help/s/question/0D54z00009RNZAICA5/self-service-approval-multiple-approvers-only-requiring-to-be-approved-by-one-of-the-approvers — 2026-11-03 — fórum
- https://www.bleepingcomputer.com/news/security/okta-we-made-a-mistake-delaying-the-lapsus-hack-disclosure — 2026-11-03 — imprensa
- https://therecord.media/sitel-blames-okta-breach-on-legacy-network-from-acquisition — 2026-11-03 — imprensa
- https://www.bleepingcomputer.com/news/security/okta-breach-134-customers-exposed-in-october-support-system-hack/ — 2026-11-03 — imprensa
- https://cloudsecurityalliance.org/articles/five-lessons-learned-from-okta-s-support-site-breach — 2026-11-03 — análise
- https://goauthentik.io/blog/2023-10-23-another-okta-breach/ — 2026-11-03 — concorrente
- https://www.computerweekly.com/news/366571853/Okta-doubles-down-on-cyber-in-wake-of-high-profile-breaches — 2026-11-03 — imprensa
- https://www.okta.com/blog/2025/01/okta-secure-identity-commitment-2024-year-in-review-highlights-innovations-and-whats — 2026-11-03 — oficial (autodeclarado)
- https://techcrunch.com/?p=2461759 — 2026-11-03 — imprensa (roubo de código no GitHub)
- https://securityboulevard.com/2024/11/saas-security-alert-ad-ldap-delegated-authentication-bypass-username-above-52-characters/ — 2026-11-03 — imprensa
- https://trust.okta.com/security-advisories/okta-ad-ldap-delegated-authentication-username/ — 2026-11-03 — oficial (não aberto; citado por fontes secundárias)
- https://services.nvd.nist.gov/rest/json/cves/2.0?keywordSearch=auth0 e ?keywordSearch=okta — 2026-11-03 — NVD
- https://news.ycombinator.com/item?id=26334516 (aquisição), 39380790 (Goodbye Auth0), 26880147 (queda), 30762520 (Lapsus$, 1158 pontos), 49478091 — 2026-11-03 — comunidade
