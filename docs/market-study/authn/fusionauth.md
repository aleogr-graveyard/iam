# FusionAuth

| Campo | Valor |
|---|---|
| Categoria | CIAM autohospedável **e** IDaaS (FusionAuth Cloud, instância dedicada) — "single-tenant por instalação, multi-tenant dentro da instância" |
| Site / repositório | https://fusionauth.io · issues públicas em https://github.com/FusionAuth/fusionauth-issues · bibliotecas abertas (ex.: `fusionauth-samlv2`, `java-http`, SDKs) em https://github.com/FusionAuth |
| Licença e modelo de negócio | **Código do servidor fechado** (não é open source): binário gratuito "Community" sob a FusionAuth Customer License Agreement (uso interno; proíbe redistribuir/usar por terceiros salvo contrato). Planos pagos por MAU/instância: Starter (a partir de US$ 162/mês, anual), Essentials e Enterprise (a partir de US$ 2.970/mês). Recursos "premium" exigem chave de licença mesmo autohospedado. Passkeys/WebAuthn exigem licença gratuita. Rodada de US$ 65 mi em nov/2023 (primeiro investimento externo) |
| Stack | Java (JVM embarcada; Java 25 na versão atual); framework web próprio **Prime MVC** e servidor HTTP próprio **`java-http`**; PostgreSQL ≥ 14 (único testado) ou MySQL ≥ 8; Elasticsearch/OpenSearch opcional para busca de usuários; lambdas em JavaScript; Kafka opcional para eventos |
| Maturidade | Origem na Inversoft; produto desde ~2018 (HN 2018–2019); release 1.70.0 em 28/out/2026, cadência de ~2 semanas a 1 mês; 139 avaliações no G2. Estrelas/top issues no GitHub: não verificado (github.com não acessível nesta sessão) |
| Acesso | 2026-11-03 |

> Limitação: `fusionauth.io` lido via `curl` (preços, licença, release notes, docs). `github.com` e buscas adicionais indisponíveis (orçamento de busca esgotado), então sinais de issues mais votadas não foram coletados.

## Resumo
- CIAM completo e API-first que roda igual no próprio servidor ou na nuvem deles; dados de cada cliente nunca misturados (instância dedicada).
- Forte em ciclo de vida B2C (famílias/COPPA, consentimento, formulários avançados, temas, lambdas) e em migração (import de hashes, hashing customizado, conector de banco legado).
- Muitos recursos de segurança (detecção de ameaças, senhas vazadas, MFA por e-mail/SMS, SCIM, DPoP) estão em planos pagos.
- Ritmo alto de correções de segurança em 2025–2026, divulgadas nas release notes, sem CVE para a maioria.
- Adições recentes para agentes de IA: Token Exchange (RFC 8693), Client ID Metadata Documents, entidade "AI Agent".

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ◐ | Senha, link mágico e código, passkeys, social (Google, Apple, Facebook, LinkedIn, Twitter etc. + provedores de games: Epic, Steam, Twitch, Xbox, Nintendo — pago), OIDC/SAML de entrada, "External JWT", LDAP (pago). ICP-Brasil: não encontrado |
| 2 | MFA e step-up | ◐ | TOTP (grátis); e-mail, SMS, voz (pagos); WebAuthn plataforma (grátis com licença) e cross-platform (pago); códigos de recuperação (hash em repouso desde 2026); políticas por tenant e por app; "Intelligent MFA" e impossible travel (pagos); lambda de MFA. Push e WhatsApp: não encontrados |
| 3 | Ciclo de vida da conta | ● | Cadastro (formulários avançados), verificação de e-mail, recuperação, vinculação de IdPs, provisionamento, bloqueio, exclusão; **famílias e consentimento (COPPA)**; SCIM server (pago); bloqueio de cadastro por domínio. Verificação de idade além de COPPA: não encontrado |
| 4 | Sessões e tokens | ◐ | JWT + refresh tokens (rotação/uso único, sliding), revogação, sessões de longa duração, gestão de sessões por dispositivo, **DPoP** (Enterprise). Back-channel logout: não encontrado (há logout URLs/front-channel) |
| 5 | Protocolos | ◐ | OAuth 2.0, OIDC, SAML v2 (SP e IdP, IdP-initiated pago), SCIM, WebAuthn, device authorization grant, Token Exchange, CIMD. CIBA, PAR, FAPI: não encontrados |
| 6 | Máquinas e agentes | ◐ | Client credentials via **Entity Management** (entidades com permissões; pago, com limites por plano), chaves de API com escopo de tenant, Token Exchange e CIMD (Enterprise) para agentes, tipo de entidade "AI Agent". Identidade de workload: não encontrado |
| 7 | Detecção e defesa | ◐ | Grátis: força bruta/bloqueio de conta. Pagos: senhas vazadas, rate limiting configurável por tenant, CAPTCHA, ACL por IP/localização, notificações de ameaça (Advanced Threat Detection = Enterprise) |
| 8 | Privacidade e conformidade | ● | Autohospedagem (soberania total), instância dedicada na nuvem com escolha de região, FIPS (pago), HIPAA BAA, GDPR/COPPA, consentimento |
| 9 | Experiência e administração | ● | Admin UI completa, temas FreeMarker (simples/avançado, versionáveis via Git e CLI), i18n de UI e e-mails, portal self-service de conta (pago), webhooks (transacionais), Kafka, relatórios (DAU/MAU, logins), API completa e SDKs, servidor MCP |
| 10 | Operação e migração | ● | Importação em massa com hashes (bcrypt, PBKDF2, SHA etc.) e plugins de hashing customizado, Connectors para migração preguiçosa (pago), Kickstart (configuração inicial declarativa), Docker/Helm, upgrades com migração de banco automática |
| 11 | Identidade descentralizada | ○ | Não encontrado |

## Diferenciais / exclusivos
- **Mesmo binário em self-host e cloud**; cloud é instância dedicada por cliente (sem co-mistura de dados).
- **Modelagem de famílias** ("supports divorce") e consentimento parental (COPPA) — raro no mercado.
- **Lambdas** JavaScript em pontos do fluxo (popular JWT, reconciliar IdP, validação de login, MFA) — com opção de chamadas HTTP (pago).
- **Kickstart**: arquivo declarativo que configura a instância na primeira inicialização (bom para CI e ambientes efêmeros).
- **Dependências mínimas por decisão**: framework web e servidor HTTP próprios (Prime MVC, `java-http`) em vez de Tomcat/Spring.
- **Agentes de IA**: Token Exchange + Client ID Metadata Documents (autorregistro de clientes, alinhado ao MCP) em 1.70.

## O que a comunidade elogia (com evidência)
- Controle dos dados e APIs simples ("control over the data is guaranteed") — G2 (https://g2.com/products/fusionauth/reviews), avaliação de usuário.
- Instalação fácil e Helm chart para Kubernetes — Software Advice.
- Suporte rápido e cordial — G2.
- Release notes muito detalhadas, inclusive de segurança — https://fusionauth.io/docs/release-notes (observação direta).

## Do que a comunidade se queixa (com evidência)
- **Preço do self-hosted pago / entrada alta** para experimentar recursos premium; estrutura de custos pouco clara em escala — G2 e Software Advice.
- **Curva de aprendizado** e UI admin difícil em cenários avançados — G2.
- **UI hospedada limitada** para branding customizado (temas FreeMarker) — G2.
- **Dependência de Elasticsearch** em versões antigas; MariaDB não suportado oficialmente — Software Advice/G2 (a doc atual confirma: só PostgreSQL é testado; MySQL com driver baixado à parte por licença).
- **Não é open source** apesar de gratuito — por inferência da licença (proíbe redistribuição); concorrentes open source exploram isso (sem evidência de discussão coletada nesta sessão).

## Histórico de segurança
| ID | Ano | Classe | Lição para nós |
|---|---|---|---|
| CVE-2020-7799 | 2020 | Injeção de template (FreeMarker) por quem edita e-mails/temas → execução de código (CVSS 7,2) | Templates editáveis por tenants precisam de sandbox estrita; repetida em 2026 ("theme_manager escapava do sandbox de templates") |
| CVE-2020-12676 | 2020 | `fusionauth-samlv2`: asserção sem `Signature` aceita (signature exclusion, CVSS 9,1) | SAML: exigir assinatura sempre; testes negativos de XML Signature (exclusão, wrapping) |
| CVE-2021-27736 | 2021 | XXE no parser SAML | Desabilitar DTD/entidades externas em todo parser XML |
| CVE-2022-45921 | 2022 | Path traversal lendo arquivos fora da raiz | Servir arquivos estáticos só de lista permitida |
| Release notes 1.54 | 2024 | Bypass de etapas pós-autenticação em páginas hospedadas (introduzido em 1.41) | Máquina de estados do fluxo de login validada no servidor a cada passo |
| Release notes 2024–2025 | 2024–25 | Correção adicional na verificação de asserções SAML cifradas/assinadas; XSS no admin; race em refresh token de uso único; enumeração de e-mails pelo "esqueci a senha"; vinculação por e-mail sem checar vínculo existente | Mesmas classes aparecem em todo IdP: SAML, enumeração, reuso de token, vinculação |
| Release notes 1.65–1.66 | 2026 | Chaves de API com escopo de tenant alcançavam endpoints globais (gerar chaves) e adicionavam IdP a app de outro tenant | Isolamento de tenant nas chaves de API deve ser testado endpoint a endpoint |
| Release notes 1.69.x–1.70 | 2026 | Vários: tomada de conta crítica (1.69.2), código de autorização reutilizável, verificação de RP no WebAuthn, DoS no `java-http`/Prime MVC, refresh tokens ilimitados → DoS, isolamento de tenant, SAML, MFA/passkey | Componentes próprios (HTTP, MVC) herdam toda a superfície de ataque; premissa 24 exige fuzzing e testes ofensivos contínuos nesses componentes |

Observação: a maioria das falhas de 2024–2026 foi divulgada só nas release notes, sem CVE/GHSA — dificulta rastreamento por scanners.

## Arquitetura e decisões notáveis
- Monólito Java com API REST completa; admin UI usa a mesma base; páginas hospedadas renderizadas no servidor (FreeMarker).
- **Single-tenant por instalação** como princípio de segurança/soberania; "Tenants" dentro da instância separam usuários, apps, grupos, configurações (SMTP, políticas, MFA, rate limits) — o mesmo e-mail pode existir em tenants diferentes.
- Banco relacional como fonte da verdade; motor de busca opcional (Elasticsearch/OpenSearch) ou busca no banco.
- Licença como "chave" que liga recursos no mesmo binário (open core fechado).
- Eventos via webhooks transacionais (a operação falha se o webhook falhar, opcional) e Kafka.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- **Multi-tenancy**: tenants nativos com isolamento lógico; chaves de API restritas a tenant; "SSO Tenant Manager" (Enterprise) para delegar administração a clientes. Não há "organizações" B2B dentro de um tenant (não encontrado; tenants cumprem esse papel).
- **Extensibilidade**: lambdas JS, webhooks, Kafka, Connectors (LDAP e API genérica), plugins Java de hashing.
- **Configuração**: admin UI, API, Kickstart, Terraform provider e CLI para temas/e-mails.
- **Bancos**: PostgreSQL (recomendado, único testado), MySQL 8+ (driver separado), sem MySQL Group Replication; busca em Elasticsearch/OpenSearch ou banco.
- **Soberania**: total no self-host (inclusive air-gapped com recursos premium); cloud com instâncias dedicadas por região.

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **não** encontrado (docs, preços, release notes).
- **Autorização sensível ao estado / transições declarativas**: **não**. Oferece RBAC (papéis por app), Entity Management (permissões entre entidades) e menciona ABAC/ReBAC e "Fine-grained Authorization Support" na tabela de preços, sem modelo de estados ou transições.

## Implicações para o projeto
1. **Adotar**: mesmo artefato para self-host e SaaS, com instância/isolamento por cliente opcional — atende soberania (premissa 9/19) e é argumento comercial.
2. **Adotar**: configuração declarativa de bootstrap (estilo Kickstart) para o usuário master e ambientes efêmeros (seção 4.23).
3. **Adotar**: importação com múltiplos algoritmos de hash + migração preguiçosa por conector — reduz barreira de entrada.
4. **Oportunidade**: ser **de fato** open source (o FusionAuth é gratuito mas fechado) e não travar segurança básica (senhas vazadas, rate limiting, CAPTCHA, MFA por e-mail) em licença.
5. **Evitar**: templates com linguagem poderosa sem sandbox (duas fugas de sandbox em 6 anos); preferir templates lógicos mínimos.
6. **Adotar com cautela**: componentes próprios no lugar de dependências (premissa 24) exigem fuzzing, testes de DoS e auditoria — o FusionAuth teve DoS no seu `java-http`/Prime MVC.
7. **Adotar**: publicar CVE/GHSA para toda correção de segurança, não só release notes (premissa 5, auditável).
8. **Adotar**: famílias/consentimento parental e Token Exchange/CIMD para agentes como itens do roadmap (domínios 3 e 6).

## Fontes
- https://fusionauth.io/pricing — 2026-11-03 — oficial
- https://fusionauth.io/license — 2026-11-03 — oficial (Customer License Agreement)
- https://fusionauth.io/feature-list — 2026-11-03 — oficial
- https://fusionauth.io/docs/release-notes — 2026-11-03 — oficial (1.70.0 de 28/10/2026)
- https://fusionauth.io/docs/get-started/download-and-install/system-requirements — 2026-11-03 — oficial
- https://fusionauth.io/docs/lifecycle/authenticate-users/oauth/endpoints — 2026-11-03 — oficial (DPoP Enterprise, device grant)
- https://fusionauth.io/docs/get-started/core-concepts/tenants — 2026-11-03 — oficial
- https://fusionauth.io/docs/extend/events-and-webhooks — 2026-11-03 — oficial
- https://fusionauth.io/docs/operate/secure/advanced-threat-detection — 2026-11-03 — oficial
- https://services.nvd.nist.gov/rest/json/cves/2.0?keywordSearch=fusionauth — 2026-11-03 — NVD (4 CVEs)
- https://g2.com/products/fusionauth/reviews — 2026-11-03 — site de avaliação
- https://www.softwareadvice.com/ciam/fusionauth-profile/reviews/ — 2026-11-03 — site de avaliação (algumas avaliações de 2020)
- https://news.ycombinator.com/item?id=38098326 (rodada US$ 65 mi) e 41265397 (1.52 passkeys) — 2026-11-03 — comunidade
