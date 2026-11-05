# Keycloak

| Campo | Valor |
|---|---|
| Categoria | AuthN, serviço self-hosted (IAM/IdP completo, com Authorization Services embutido) |
| Site / repositório | https://www.keycloak.org · https://github.com/keycloak/keycloak |
| Licença e modelo de negócio | Apache-2.0, sem edição "enterprise" aberta/fechada: todo o produto é aberto. Monetização indireta: Red Hat vende suporte do downstream "Red Hat build of Keycloak" (antes RH-SSO). Ecossistema de terceiros vende hospedagem/extensões (Phase Two, Skycloak, Cloud-IAM). Sem mudança de licença |
| Stack | Java (Quarkus desde a v17; antes WildFly), Infinispan para cache/sessões distribuídas, Hibernate/JPA. Bancos: PostgreSQL (inclui Aurora PostgreSQL), MySQL, MariaDB, MS SQL Server, Oracle; TiDB citado no guia; `dev-file` só para desenvolvimento. Temas em FreeMarker; consoles em React (PatternFly) |
| Maturidade | ~2014 (projeto JBoss/Red Hat). CNCF Incubating desde abr/2023. ~37,7 mil estrelas, ~9,1 mil forks. Release 26.8.1 em 13/out/2026; cadência de minor ~trimestral com patches quinzenais. Adoção ampla (ecossistema Red Hat/OpenShift, governo, academia) |
| Acesso | 2026-11-03 |

> Limitação: `keycloak.org` e `news.ycombinator.com` não resolviam na rede da sessão. Usei GitHub (páginas, advisories, clone esparso da documentação em `docs/documentation` e `docs/guides`), API Algolia do HN e buscas na web.

## Resumo
- IdP open source mais completo e mais adotado: OIDC/OAuth/SAML, brokering, federação LDAP/Kerberos, FAPI 2.0, CIBA, PAR, DPoP, Organizations, SCIM (supported em 26.8).
- 100% Apache-2.0, governança CNCF; inovação rápida em padrões (OID4VCI/OID4VP, AuthZEN, delegação para agentes de IA, MCP).
- Extensível por SPIs Java; customização de UI via temas FreeMarker.
- Queixas recorrentes: peso (JVM, memória), curva de aprendizado, documentação, temas que quebram em upgrades, escala por número de realms.
- Volume alto de advisories (~90 no GitHub), muitos em redirect/brokering/SAML/DCR.

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ◐ | Senha, passkeys com UI condicional e modal (26.4), login social, brokering OIDC/SAML, federação LDAP/AD e Kerberos, certificado X.509 (base para ICP-Brasil, sem perfil ICP específico). Link mágico e código por e-mail sem suporte nativo (extensões de terceiros; "magic link" ausente da doc) |
| 2 | MFA e step-up | ◐ | TOTP/HOTP, WebAuthn (chaves e plataforma), códigos de recuperação (oferta acoplada ao OTP em 26.4), step-up por LoA/ACR (inclusive SAML em 26.7). SMS, WhatsApp, push e MFA adaptativo por risco: não nativos (SPI) |
| 3 | Ciclo de vida da conta | ◐ | Cadastro, verificação de e-mail, reset, user profile declarativo, vinculação de IdP (first broker login), convites de organização e via Workflows (26.8), JIT por brokering, SCIM API supported (26.8), exclusão pelo usuário. Bloqueio via brute force. Verificação de idade: não encontrado. SCIM de saída (cliente): não encontrado |
| 4 | Sessões e tokens | ● | Sessões por dispositivo no Account Console, revogação, rotação de refresh token, back-channel e front-channel logout, DPoP supported (26.4), sessões offline, persistência de sessões em banco |
| 5 | Protocolos | ● | OAuth 2.0 (perfis de client policy para OAuth 2.1), OIDC, SAML 2.0, SCIM 2.0, WebAuthn, device flow, CIBA, PAR, FAPI 1 e FAPI 2.0 (Security Profile e Message Signing finais em 26.4), JWT Authorization Grant RFC 7523 (26.6), Client ID Metadata Document (experimental), Shared Signals (experimental) |
| 6 | Máquinas e agentes | ◐ | Client credentials, service accounts, token exchange (padrão), federated client authentication (26.6), guia de authorization server para MCP, delegação por token exchange com claim `act` (preview em 26.8). Chaves de API para usuários finais: não nativo |
| 7 | Detecção e defesa | ◐ | Detecção de brute force (falhas agora persistidas em banco, 26.8), reCAPTCHA no cadastro, lista de senhas proibidas (recarregável em 26.6). Verificação de senhas vazadas (HIBP), bot detection e anomalias: não nativos. Rate limiting: delegar a proxy |
| 8 | Privacidade e conformidade | ◐ | Consentimento OAuth, exclusão de conta, user profile com permissões por atributo, eventos de auditoria; modo FIPS. Exportação de dados pelo titular: não encontrado como recurso nativo |
| 9 | Experiência e administração | ◐ | Admin Console e Account Console (React), temas FreeMarker (Keycloakify como alternativa comunitária), i18n com RTL (26.6), Admin REST API v2 declarativa (preview em 26.8). Webhooks nativos: não ("webhook" ausente da doc); eventos só via SPI de event listener |
| 10 | Operação e migração | ● | Operator Kubernetes, Helm (experimental), import/export de realm, atualizações de patch sem downtime (26.6), multi-cluster v2 supported (26.8), federação LDAP e SPI de storage para migração gradual |
| 11 | Identidade descentralizada | ◐ | OID4VCI em preview com revogação (26.8); verificador OID4VP experimental (26.8). Issue "Log in using the EUDI-Wallet" com 96 votos |

## Diferenciais / exclusivos
- **Amplitude de padrões em OSS puro**: FAPI 2.0, CIBA, PAR, DPoP, OID4VCI/VP, Shared Signals, AuthZEN, sem edição paga.
- **Authorization Services** (UMA 2.0, políticas por recurso/escopo, estratégias de decisão unânime/afirmativa/consenso) e **AuthZEN PDP** experimental em 26.7, com avaliação individual e em lote.
- **Fine-Grained Admin Permissions v2** para delegação administrativa, inclusive por organização (26.7).
- **Workflows** (26.6 supported): automação declarativa (YAML) do ciclo de vida disparada por eventos ou agenda.
- **Delegação para agentes de IA**: scopes `delegation:client:<id>` com consentimento e claim `act`, sem escalar a Admin API.
- **Federação de usuários** (LDAP/AD, Kerberos, SPI de user storage) para migração sem mover dados.

## O que a comunidade elogia (com evidência)
- Completo, estável e flexível em produção há anos, com muitos IdPs integrados ([HN 31258469](https://hn.algolia.com/api/v1/items/31258469), usuários `nird`, `sz4kerto`, `maxbaines`).
- Configuração muito mais simples após Quarkus, por variáveis de ambiente (`sascha_sl`, mesmo thread).
- API administrativa útil para provisionar clientes e provedores no deploy (`imglorp`, mesmo thread).
- Ritmo de inovação em padrões (release notes 26.4–26.8 no repositório).

## Do que a comunidade se queixa (com evidência)
- **Peso**: "a loooooooot of memory", imagens grandes; ~2 GiB de RAM inviável em VPS pequena; tratado como "pet", não "cattle" ([HN 31258469](https://hn.algolia.com/api/v1/items/31258469): `psankar`, `CGamesPlay`, `RandomBK`). Relato Red Hat de OOM ao migrar de 25.0.6 para 26.4.1 com ~170 realms (issue RHBK-3932, via busca).
- **Documentação e curva de aprendizado**: "documentation is poor", guias desatualizados, difícil sem dominar OAuth/OIDC (`pharmakom`, `schipplock`, `doctor_eval`, mesmo thread).
- **Customização da UI**: temas exigem stack Java; difícil para dev front-end (`littlecranky67`, mesmo thread). Temas e extensões quebram em upgrades (guia de upgrade 26.8 cita botões de IdP redesenhados).
- **Demora em correções de segurança**: análise de 2024 relata até ~10 meses entre reporte e correção ([humanativaspa](https://security.humanativaspa.it/an-analysis-of-the-keycloak-authentication-system/); [HN 42136000](https://hn.algolia.com/api/v1/items/42136000)).
- **Lacunas antigas mais votadas** ([issues por reações](https://github.com/keycloak/keycloak/issues?q=is%3Aissue+is%3Aopen+sort%3Areactions-%2B1-desc)): SCIM #13484 (242, aberta desde 2022 apesar do SCIM nativo recente), dispositivo confiável #8742 (112), Organizations #30180 (99), CSP endurecida #16277 (79), segredos via arquivo #10816 (47), clonar realm #24770.

## Histórico de segurança
~90 advisories no GitHub (9 páginas em 2026-11-03). Amostra:

| ID | Ano | Classe | Lição |
|---|---|---|---|
| GHSA-j97h-3f8r-mrjr (CVE-2026-11800) | 2026 | Confusão de algoritmo JWT no JWT Authorization Grant (CWE-347) | Fixar algoritmos por chave/cliente; nunca aceitar `alg` do token |
| GHSA-xpwp-2pcm-8xq3 | 2026 | Bypass de proteção contra replay por diferença de semântica do driver de banco | Testar invariantes de segurança em todos os bancos suportados |
| GHSA-95cx-vmr5-3cmr, GHSA-95rm-h7g9-rhcf | 2026 | DCR: forja de papéis via mappers; troca de tipo de mapper burla política | Registro dinâmico de cliente é superfície crítica; política deny-by-default |
| GHSA-f8m4-v488-rmrm, GHSA-fgq2-hxm5-8xg2 | 2026 | SAML broker: import de metadata desliga validação de assinatura; login IdP-initiated burla "link-only" | Import de configuração não pode rebaixar segurança; testar fluxos raros |
| GHSA-2888-g6qc-w4mj, GHSA-f5p5-6xmx-p252 | 2026 | Bypass de autorização por URI não normalizada | Normalizar/canonicalizar antes de comparar |
| GHSA-hmr6-pxx9-552p | 2026 | Busca LDAP por entry-dn ignora base DN | Validar fronteiras em conectores |
| GHSA-27gp-8389-hm4w | 2025 | Escalada de privilégio no Admin Console com FGAP v2 | Delegação administrativa precisa de testes de autorização dedicados |
| GHSA-5jfq-x6xp-7rw2 | 2025 | Bypass de 2FA | Cobertura de testes de todos os caminhos de fluxo |
| GHSA-2p82-5wwr-43cw | 2025 | Bypass por falta de bind LDAP após reset de senha | Revalidar credencial na fonte após mudanças |
| GHSA-mrv8-pqfj-7gp5, GHSA-mpwq-j3xf-7m5w, GHSA-cvg2-7c3j-g36j | 2023–2024 | `redirect_uri`: path traversal, bypass de hosts, XSS com curinga | Comparação exata de redirect URI (OAuth 2.1); proibir curingas |
| GHSA-5q66-v53q-pm35 | 2023 | Senha em texto claro armazenada | Revisar todo caminho de persistência de segredos |
| GHSA-8hxp-qmph-w5gq | 2025 | Resolução de variáveis em import expõe env vars | Interpolação em configuração é superfície de ataque |

## Arquitetura e decisões notáveis
- Monólito Java/Quarkus com build otimizado ("augmentation") e cache distribuído Infinispan; sessões agora persistidas em banco; multi-cluster v2 sem caches externos (26.7/26.8).
- Extensibilidade por SPIs Java (autenticadores, storage, mappers, event listeners, temas) empacotados como JARs no servidor: poderoso, porém acopla upgrades.
- Fluxos de autenticação configuráveis como árvore de execuções (required/alternative/conditional).
- Recursos em estágios explícitos: experimental → preview → supported, ligados por feature flags (`--features=`).

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- **Multi-tenancy**: realms (isolamento forte, mas custo por realm; relatos de limite de escala e memória com centenas de realms) + Organizations (26.x) para B2B dentro de um realm, com grupos de organização, IdPs compartilhados e admin delegado.
- **Extensibilidade**: SPIs Java; Workflows declarativos; sem hooks HTTP nativos.
- **Configuração**: CLI/env/arquivo para o servidor; realm via Admin API, import/export JSON; Admin API v2 declarativa (preview); provider Terraform comunitário.
- **Bancos**: PostgreSQL (e Aurora), MySQL, MariaDB, MSSQL, Oracle; TiDB citado.
- **Soberania**: self-hosted, sem dependência de nuvem; ocupa um realm por jurisdição ou um cluster por região (sem roteamento de dados por tenant nativo).

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **não**. Authorization Services combina políticas automáticas (estratégias unânime/afirmativa/consenso), não aprovações humanas; Workflows automatiza ciclo de vida, sem etapa de aprovação por quórum (release notes 26.6–26.8 e `server_admin`, buscas por "approval"). Não encontrado.
- **Autorização sensível ao estado / transições declarativas**: **não**. Políticas por recurso/escopo (papel, grupo, cliente, tempo, usuário, agregadas), sem modelo de estados do recurso nem de transições. Não encontrado.

## Implicações para o projeto
1. **Adotar como barra mínima de protocolos**: FAPI 2.0, PAR, DPoP, CIBA, device flow, back-channel logout, SCIM; o OSS de referência já tem tudo sem edição paga.
2. **Adotar estágios de maturidade explícitos** (experimental/preview/supported) com feature flags, alinhado a "alertar em vez de bloquear".
3. **Evitar** extensibilidade por plugins in-process acoplados à versão; preferir hooks/webhooks versionados e contratos estáveis.
4. **Oportunidade**: leveza (binário Go, baixa RAM, boot rápido) e multi-tenancy barato por tenant, onde realms escalam mal.
5. **Oportunidade**: UI de login customizável sem stack de backend (templates/headless + componentes), queixa recorrente.
6. **Segurança**: matriz de testes para redirect URI, DCR, import de metadata SAML e normalização de URIs; política de SLA público de correção (queixa de 10 meses).
7. **Acompanhar**: AuthZEN no IdP e delegação para agentes (`act`), sinais de que AuthN e AuthZ convergem — reforça a família de produtos.

## Fontes
- https://github.com/keycloak/keycloak — 2026-11-03
- https://github.com/keycloak/keycloak/releases — 2026-11-03
- https://github.com/keycloak/keycloak/security/advisories (páginas 1, 3 e 6) — 2026-11-03
- https://github.com/keycloak/keycloak/security/advisories/GHSA-j97h-3f8r-mrjr — 2026-11-03
- https://github.com/keycloak/keycloak/issues?q=is%3Aissue+is%3Aopen+sort%3Areactions-%2B1-desc — 2026-11-03
- Repositório, `docs/documentation/release_notes/topics/26_6_0.adoc`, `26_7_0.adoc`, `26_8_0.adoc`, `27_0_0.adoc`; `docs/documentation/server_admin`; `docs/guides/server/db.adoc` (clone esparso) — 2026-11-03
- https://www.keycloak.org/2025/09/keycloak-2640-released (via busca; passkeys, DPoP, FAPI 2.0) — 2026-11-03
- https://www.keycloak.org/2026/04/scim-as-experimental-feature (via busca) — 2026-11-03
- https://www.keycloak.org/2025/10/workflows-experimental-26-4 (via busca) — 2026-11-03
- https://cncf.io/blog/2023/04/11/keycloak-joins-cncf-as-an-incubating-project — 2026-11-03
- https://hn.algolia.com/api/v1/items/31258469 (HN "Keycloak: Open-Source IAM", 2022, 388 pontos) — 2026-11-03 — opiniões individuais; comentários de concorrentes (FusionAuth, Zitadel) identificados
- https://hn.algolia.com/api/v1/items/42136000 (HN, 2024) — 2026-11-03
- https://security.humanativaspa.it/an-analysis-of-the-keycloak-authentication-system/ — 2026-11-03 — pesquisa de segurança independente (não lida diretamente; resumo via HN)
- https://skycloak.io/blog/keycloak-native-scim-api-preview-26-7/ — 2026-11-03 — viés: vendedor de hospedagem Keycloak
- https://hoop.dev/blog/the-keycloak-pain-point/ e https://www.authgear.com/post/keycloak-upgrade-guide (via busca) — 2026-11-03 — viés: concorrentes/fornecedores
- https://issues.redhat.com/browse/RHBK-3932 (via busca; OOM 25→26) — 2026-11-03
