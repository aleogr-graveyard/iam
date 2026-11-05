# Ory (Kratos, Hydra, Oathkeeper, Polis)

| Campo | Valor |
|---|---|
| Categoria | AuthN, serviços self-hosted "headless" e componíveis (identidade, servidor OAuth2/OIDC, proxy de acesso, ponte SAML/SCIM) + SaaS (Ory Network) |
| Site / repositório | https://www.ory.com · https://github.com/ory/kratos · https://github.com/ory/hydra · https://github.com/ory/oathkeeper · https://github.com/ory/polis (AuthZ: Ory Keto, estudado no grupo AuthZ) |
| Licença e modelo de negócio | OSS em Apache-2.0 (Kratos, Hydra, Oathkeeper, Polis). **Open core**: a Ory Enterprise License (OEL) adiciona recursos fechados (SAML, SCIM, login por organização/B2B SSO, CAPTCHA, multi-tenancy, escala avançada), builds em registro Docker privado e **patches de CVE com SLA** (crítico em 14 dias, alto em 30, segundo política do fosite). O README do Kratos diz que o OSS serve para "unimportant workloads without SLAs". Ory Network: SaaS com preço por uso e plano gratuito de desenvolvedor. Sem mudança da licença OSS; a mudança foi de **modelo** (recursos e garantias migrando para OEL) |
| Stack | Go. Bancos: PostgreSQL, MySQL, CockroachDB (SQLite para desenvolvimento); Hydra baseado na biblioteca `ory/fosite`. Polis (ex-BoxyHQ SAML Jackson) em TypeScript/Next.js, com bancos "traga o seu" (Postgres, MySQL, MariaDB, MongoDB, Redis, PlanetScale etc.) |
| Maturidade | Empresa alemã, projetos desde ~2015 (Hydra). Estrelas: Hydra ~17,6 mil; Kratos ~13,9 mil; Oathkeeper ~3,6 mil; Polis ~2,3 mil. Kratos v1.0 em jul/2023; versionamento passou a calendário (v25.4.0, v26.2.0 em 20/mar/2026). Hydra OpenID Certified. Adotantes listados no README do Hydra: OpenAI, Mistral, Cisco, Klarna, Raspberry Pi, Arduino, entre outros. BoxyHQ adquirida em mai/2025 → Ory Polis |
| Acesso | 2026-11-03 |

> Limitação: `www.ory.com`, `www.ory.sh` e `news.ycombinator.com` não resolviam na rede da sessão. Usei GitHub (READMEs, releases, advisories, issues, `embedx/config.schema.json` do Kratos e `.schema/config.schema.json` do Hydra via raw.githubusercontent.com), API Algolia do HN e buscas na web.

## Resumo
- Suíte de microsserviços: Kratos (identidade e fluxos de autoatendimento), Hydra (OAuth2/OIDC), Oathkeeper (proxy/decisão de acesso), Polis (SAML→OAuth e SCIM), Keto (ReBAC estilo Zanzibar).
- Totalmente "headless": APIs e fluxos; UI fica com o integrador (Ory Elements como kit).
- Configuração declarativa em YAML com JSON Schema; identidades descritas por JSON Schema próprio.
- Open core cada vez mais pronunciado: SAML, SCIM, B2B SSO, CAPTCHA, multi-tenancy e patches de CVE com SLA só na OEL/Ory Network.

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ◐ | Kratos (métodos no schema de config): `password`, `code` (sem senha por e-mail/SMS), `passkey`, `webauthn`, `oidc` (social e OIDC genérico), `link` (link mágico, legado para recuperação/verificação). SAML/B2B SSO: só OEL/Ory Network (`b2b`: "not available in the open source licensed code"); Polis OSS faz ponte SAML→OAuth. LDAP/AD e certificado ICP-Brasil: não encontrados |
| 2 | MFA e step-up | ◐ | TOTP, WebAuthn, `lookup_secret` (códigos de recuperação), `code` por e-mail/SMS como 2º fator; níveis AAL1/AAL2 com `required_aal` e `highest_available` (step-up). Push, WhatsApp e MFA adaptativo: não encontrados |
| 3 | Ciclo de vida da conta | ◐ | Fluxos de cadastro, login, verificação, recuperação e configurações; vinculação de contas OIDC; Admin API de identidades (criação, importação com hashes, exclusão). Convites ausentes (issue #2525, aberta desde 2022). SCIM: OEL ou Polis (Directory Sync). JIT: via OIDC. Bloqueio por tentativas: não encontrado no OSS |
| 4 | Sessões e tokens | ◐ | Kratos: sessões por dispositivo, listagem/revogação, `whoami` com "tokenizer" para JWT. Hydra: rotação de refresh token com período de tolerância, revogação, front-channel e back-channel logout. DPoP: não encontrado nos schemas |
| 5 | Protocolos | ◐ | Hydra: OAuth 2.0/2.1, OIDC certificado (Basic, Implicit, Hybrid, Discovery, Dynamic), DCR, device authorization grant (v25.4), JWT bearer grant. WebAuthn (Kratos). SAML e SCIM via Polis/OEL. PAR, CIBA e FAPI: não encontrados (schema do Hydra sem PAR) |
| 6 | Máquinas e agentes | ◐ | Hydra: client credentials, `private_key_jwt`, JWT bearer; device grant posicionado para "agentic authorization". Oathkeeper autentica requisições (JWT, introspecção, cookie de sessão). Chaves de API para usuários finais e identidade de workloads: não encontrados |
| 7 | Detecção e defesa | ◐ | Checagem de senha vazada via HIBP **ligada por padrão** (`haveibeenpwned_enabled`, host configurável), mitigação de enumeração de contas (desligada por padrão; issue #133 aberta desde 2019). CAPTCHA: OEL. Rate limiting: Ory Network/proxy |
| 8 | Privacidade e conformidade | ◐ | Self-hosted dá controle total dos dados; Ory Network declara armazenamento "GDPR friendly" com localidade. Exportação para titular: não encontrado |
| 9 | Experiência e administração | ◐ | Headless; Ory Elements (componentes de UI). Webhooks (`web_hook`) em pontos dos fluxos; courier de e-mail/SMS com templates. Painel administrativo: só Ory Console (Network); no OSS não há UI admin |
| 10 | Operação e migração | ◐ | Migrações SQL por CLI, importação de identidades com hashes de outros IdPs; guia de migração do Auth0/Okta (Hydra + Kratos). Sem SLA de patches no OSS |
| 11 | Identidade descentralizada | ○ | Não encontrado |

## Diferenciais / exclusivos
- **Separação radical de responsabilidades**: identidade (Kratos), emissão de tokens (Hydra) e enforcement (Oathkeeper) como serviços independentes; Hydra delega login/consentimento a qualquer app ("login and consent app").
- **Identity schema em JSON Schema**: traits customizáveis, com marcação de quais campos são identificadores, verificáveis ou recuperáveis.
- **Fluxos de autoatendimento como API** com suporte explícito a navegador e apps nativos (mesma máquina de estados).
- **Níveis de garantia (AAL)** como conceito de primeira classe nas sessões.
- **HIBP por padrão**, com host próprio para soberania de dados.

## O que a comunidade elogia (com evidência)
- Escala e adoção em produção grande (README do Hydra lista OpenAI, Klarna, Cisco etc. — viés: fornecedor).
- Bom para quem quer controle total da UI e da arquitetura: "Works like a dream for the most part" (`Rodeoclash`, [HN 39465740](https://hn.algolia.com/api/v1/items/39465740), 2024, 203 pontos); frontend em Elixir foi "quite a pleasant developer experience" (`manuel2258`, [HN 36708907](https://hn.algolia.com/api/v1/items/36708907)).
- Lançamentos de passwordless, SMS e passkeys bem recebidos ([HN 39465740](https://hn.algolia.com/api/v1/items/39465740), [HN 40596687](https://hn.algolia.com/api/v1/items/40596687)).

## Do que a comunidade se queixa (com evidência)
- **Complexidade**: "Ory Kratos is so complicated. Authentik is much simpler" (`v3ss0n`, HN 39465740); "quite cool, but quite hard" (HN 36708907); "não é alternativa ao Auth0, é um componente de uma" (`rad_gruchalski`, HN 39465740).
- **UI customizada dolorosa** (`Rodeoclash`, HN 39465740); sem front-end/admin UI no self-hosted (threads do arquivo Slack da Ory, via busca).
- **Multi-tenancy ausente no OSS**: dúvida recorrente (`atonse`, HN 36708907); a doc oficial manda para Ory Network ou OEL (guia "multi-tenancy", via busca).
- **Open core e segurança**: usuário questiona a OEL porque o SDK OSS não foi atualizado por ~6 meses após correção de segurança; resposta da Ory: prazos são garantidos pela OEL ([arquivo Slack](https://archive.ory.sh/t/25972487/how-are-we-to-take-ory-entreprise-licensing-seriously-if-we-)). README classifica o OSS como para cargas sem importância.
- **Cadência OSS irregular**: Kratos ficou ~13 meses sem release (v1.3.1 out/2024 → v25.4.0 nov/2025) ([releases](https://github.com/ory/kratos/releases)).
- **Issues antigas** ([por reações](https://github.com/ory/kratos/issues?q=is%3Aissue+is%3Aopen+sort%3Areactions-%2B1-desc)): reenviar e-mail de verificação #2892, convites #2525, WebAuthn e AAL2 #2794, exigir senha atual para trocar senha #2586, enumeração #133 (2019), código de login retornando 400 #4052.
- **SMS como fator** criticado (HN 39465740, `DaiPlusPlus`), embora defendido como necessidade de mercado (`arekkas`, CEO da Ory — viés).

## Histórico de segurança
Poucos advisories, concentrados em uma divulgação coordenada em 20/mar/2026.

| ID | Ano | Produto | Classe | Lição |
|---|---|---|---|---|
| GHSA-hgx2-28f8-6g2r (CVE-2026-33503) | 2026 | Kratos | SQL injection via token de paginação forjado; tokens cifrados com **chave padrão publicamente conhecida** se `secrets.pagination` não for definida | Nenhum segredo com valor padrão; falhar ao iniciar sem segredo; não confiar em tokens opacos sem MAC |
| GHSA-r9w3-57w2-gch2 (CVE-2026-33504) | 2026 | Hydra | Mesma classe (paginação) | Componentes compartilhados propagam falhas por toda a suíte |
| GHSA-p224-6x5r-fjpm (CVE-2026-33494, CVSS 10) | 2026 | Oathkeeper | Path traversal: regra avaliada sobre caminho não normalizado (`/public/../admin`) | Normalizar antes de casar regras; enforcement e upstream devem ver o mesmo caminho |
| GHSA-4mq7-pvjg-xp2r | 2026 | Oathkeeper | Bypass por confusão de chave de cache | Chave de cache deve incluir todo o contexto de decisão |
| GHSA-vhr5-ggp3-qq85 | 2026 | Oathkeeper | Bypass por cabeçalho não confiável | Lista explícita de proxies confiáveis |
| GHSA-wc43-73w7-x2f5 | 2024 | Kratos | `highest_available` não respeitava code + MFA | Testar a matriz de combinações de fatores/AAL |
| GHSA-qvp4-rpmr-xwrr | 2021 | Oathkeeper | Bypass de validação de claims com cache de introspecção | Cache não pode pular validação |
| GHSA-3p3g-vpw6-4w66 | 2020 | Hydra | Replay de `private_key_jwt` (JTI não bloqueado) | Registrar JTI para impedir replay |

Política: no OSS, patches só na release corrente e sem SLA; OEL com SLA ([política de segurança](https://www.ory.com/docs/ecosystem/security), via busca).

## Arquitetura e decisões notáveis
- Microsserviços Go stateless sobre SQL, cada um com API pública e API admin separadas (portas distintas).
- Hydra não guarda usuários: delega autenticação e consentimento a uma aplicação externa por redirecionamentos (desacopla IdP de emissor de tokens).
- Kratos modela fluxos (login, registro, recuperação etc.) como recursos com ID, expiração e UI nodes retornados pela API; UI renderiza os nós.
- Configuração de serviço em arquivo único validado por JSON Schema publicado; recarga a quente parcial.
- Divisão OSS/OEL feita no próprio schema ("Only effective in the Ory Network or with a valid license").

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- **Multi-tenancy**: não suportada no OSS ("data model does not support this", guia oficial via busca); alternativa é uma instância por tenant. Ory Network: workspaces → projetos isolados; OEL promete multi-tenancy.
- **Extensibilidade**: webhooks em pontos de fluxo (antes/depois), identity schemas, tokenizer de sessão para JWT, Jsonnet para mapear claims de provedores OIDC.
- **Configuração**: YAML/env validados por JSON Schema; recursos via Admin API e Ory CLI.
- **Bancos**: PostgreSQL, MySQL, CockroachDB; SQLite (dev). Pedido de TiDB experimental (#4505).
- **Soberania**: self-hosted total; Ory Network com localidade de dados declarada.

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **não**. Nada nos schemas de Kratos/Hydra nem nos READMEs; Oathkeeper decide por regra, sem aprovação humana.
- **Autorização sensível ao estado / transições**: **não** no AuthN. Oathkeeper casa URL/método; Keto é ReBAC sem modelo de estados (detalhe no estudo de AuthZ). Não encontrado.

## Implicações para o projeto
1. **Adotar** fluxos de autenticação como recursos de API com máquina de estados explícita, servindo navegador e apps nativos (premissa 15).
2. **Adotar** AAL como conceito central de sessão e step-up, alinhado à premissa 2 (níveis de garantia) e ao NIST SP 800-63B.
3. **Adotar** HIBP ligado por padrão com endpoint configurável (soberania) e esquema de identidade declarativo.
4. **Evitar** segredos com valor padrão e tokens de paginação sem integridade (CVE-2026-33503); alertas de configuração (seção 4.6) devem barrar ou sinalizar isso.
5. **Evitar** open core que tira do OSS SAML, SCIM, multi-tenancy e patches de segurança: é a principal queixa e uma oportunidade de posicionamento (premissa 11).
6. **Oportunidade**: "headless" é poderoso, mas exige UI pronta e admin UI no OSS para não repetir a queixa de complexidade.
7. **Considerar** separar emissor OAuth (estilo Hydra) do núcleo de identidade, permitindo uso com IdPs de terceiros (premissa 8).

## Fontes
- https://github.com/ory/kratos (README) — 2026-11-03 — viés: fornecedor
- https://raw.githubusercontent.com/ory/kratos/master/embedx/config.schema.json — 2026-11-03
- https://raw.githubusercontent.com/ory/hydra/master/.schema/config.schema.json e README — 2026-11-03
- https://github.com/ory/hydra · https://github.com/ory/oathkeeper · https://github.com/ory/polis — 2026-11-03
- https://github.com/ory/kratos/releases — 2026-11-03
- https://github.com/ory/kratos/security/advisories · https://github.com/ory/hydra/security/advisories · https://github.com/ory/oathkeeper/security/advisories — 2026-11-03
- https://github.com/ory/kratos/security/advisories/GHSA-hgx2-28f8-6g2r · https://github.com/ory/oathkeeper/security/advisories/GHSA-p224-6x5r-fjpm — 2026-11-03
- https://github.com/ory/kratos/issues?q=is%3Aissue+is%3Aopen+sort%3Areactions-%2B1-desc — 2026-11-03
- https://www.ory.com/docs/ecosystem/security (via busca) — 2026-11-03 — viés: fornecedor
- https://github.com/ory/fosite/security/policy (via busca; SLA OEL 14/30 dias) — 2026-11-03
- https://www.ory.com/docs/kratos/guides/multi-tenancy-multitenant (via busca) — 2026-11-03 — viés: fornecedor
- https://www.ory.com/blog/introducing-ory-polis-for-enterprise-single-sign-on (via busca) — 2026-11-03 — viés: fornecedor
- https://changelog.ory.com/announcements/ann_Dne2mookJyqgU (via busca; Hydra v25.4 device grant) — 2026-11-03
- https://archive.ory.sh/t/25972487/how-are-we-to-take-ory-entreprise-licensing-seriously-if-we- (via busca) — 2026-11-03
- https://hn.algolia.com/api/v1/items/39465740 (HN 2024, 203 pontos), https://hn.algolia.com/api/v1/items/36708907 (HN 2023, Kratos v1.0) — 2026-11-03 — inclui comentários de funcionários da Ory e de concorrentes (FusionAuth, WorkOS, Stytch, authentik)
