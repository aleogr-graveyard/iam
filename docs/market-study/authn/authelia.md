# Authelia

| Campo | Valor |
|---|---|
| Categoria | AuthN, serviço self-hosted e **proxy de autenticação** (forward auth para reverse proxies) + provedor OIDC |
| Site / repositório | https://www.authelia.com · https://github.com/authelia/authelia |
| Licença e modelo de negócio | Apache-2.0, 100% comunitário, sem edição paga nem empresa por trás; financiado por doações (site pede doações para "financiar uma auditoria de segurança"). Sem mudança de licença encontrada |
| Stack | Go (servidor), React (portal). Storage: SQLite, MySQL/MariaDB, PostgreSQL (com criptografia por coluna via `storage.encryption_key`). Sessões: memória ou Redis. Usuários: arquivo YAML ou LDAP/AD. Notificação: SMTP ou arquivo |
| Maturidade | Repositório desde dez/2016. ~29,5 mil estrelas, ~1,5 mil forks, ~140 issues abertas, push diário. Versão 4.39.x (4.39.21 em 03/set/2026); v4.40 em desenvolvimento. "OpenID Certified" (mai/2025). Muito usado em homelab com Traefik/nginx/Caddy |
| Acesso | 2026-11-03 |

> Limitação: WebFetch não resolvia authelia.com, mas curl sim; usei páginas oficiais via curl. GitHub web e parte da API com rate limit; Reddit inacessível. O post "Important Announcement: The Future of Authelia" (HN 44018802, mai/2025) hoje redireciona para o post de certificação OpenID; conteúdo original não verificado.

## Resumo
- "Extensão do reverse proxy": o proxy consulta o Authelia a cada requisição; o portal emite cookie de sessão para o domínio inteiro (multi-domínio suportado).
- Política de acesso por regras (domínio/regex, sujeito, rede, método, recurso) com níveis `bypass`, `one_factor`, `two_factor`, `deny`.
- 2FA: TOTP, WebAuthn/passkeys, Duo push. Configuração inteira em YAML (infra como código), sem UI de administração.
- OIDC Provider ainda rotulado **beta** (Beta 8 em andamento), mas certificado; implementa PAR, device code, JARM, client credentials.
- Pioneiro em criptografia pós-quântica: ML-DSA em WebAuthn e JOSE desde 4.39.21.

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ◐ | Senha (arquivo YAML com hash ou LDAP/AD), passkeys sem senha (4.39.0). Login social e OIDC/SAML como RP (entrada): ausentes (issue #2728, 61 reações, aberta). SAML: ausente (#493). Magic link e ICP: não |
| 2 | MFA e step-up | ◐ | TOTP, WebAuthn (várias credenciais desde 4.38), Duo push. SMS/e-mail como fator: não. Códigos de recuperação: não encontrado. Política por regra exige 1FA/2FA; AMR (RFC 8176) e políticas por método no roadmap "Granular Authorization" |
| 3 | Ciclo de vida da conta | ◐ | Reset de senha self-service, verificação de identidade por e-mail para registrar dispositivos, "elevated session" por código único. Cadastro, convites, SCIM, exclusão: ausentes. Painel de usuários/admin no roadmap (#303, aberta desde 2018) |
| 4 | Sessões e tokens | ◐ | Cookie de sessão por domínio, Redis para HA, refresh tokens OIDC. Back-channel/front-channel/RP-initiated logout e revogação de tokens no logout: "not started" (Beta 9). DPoP: não encontrado |
| 5 | Protocolos | ◐ | OIDC (certificado), OAuth2 com PKCE por cliente, PAR, device code, JARM, JWE, client credentials, `private_key_jwt`. Token exchange, CIBA, FAPI 2.0, DCR: planejados/"not started". SAML e SCIM: ausentes |
| 6 | Máquinas e agentes | ◐ | Client credentials; autenticação por cabeçalho `Authorization` (Basic) no proxy. Tokens criados pelo usuário (#2023 fechada). Agentes de IA: não encontrado |
| 7 | Detecção e defesa | ◐ | "Regulation" (banimento temporário após N falhas), política de senha (zxcvbn/padrão). Captcha, senhas vazadas, anomalias: não encontrados |
| 8 | Privacidade e conformidade | ◐ | Criptografia de colunas sensíveis no banco, aceite de política de privacidade configurável, threat model publicado. Exportação/exclusão de dados: não |
| 9 | Experiência e administração | ○ | Portal de login com logo/tema e i18n; sem painel admin (gestão por YAML/LDAP/CLI). Webhooks: não encontrado |
| 10 | Operação e migração | ◐ | Stateless com Redis + SQL externo para HA, Helm, Kubernetes docs, binário único leve. Recarga de config ao vivo pedida (#1275, aberta desde 2020) |
| 11 | Identidade descentralizada | ○ | Não encontrado |

## Diferenciais / exclusivos
- **Forward auth como cidadão de primeira classe** com suporte documentado a vários proxies e **multi-domain protection**.
- **Cripto pós-quântica** (ML-DSA-44/65/87) para WebAuthn e assinaturas JOSE, documentada (4.39.21).
- **Regras de acesso declarativas** combinando sujeito (usuário/grupo/OAuth client), rede, método HTTP, regex de recurso e nível de autenticação.
- **Roadmap público por estágios** do OIDC, com breaking changes anunciados por versão.
- Leveza: binário Go, SQLite opcional, roda em Raspberry Pi.

## O que a comunidade elogia (com evidência)
- Simples com LLDAP; docs suficientes — HN 40952107 (2024).
- Funciona bem com nginx proxy manager e entrega identidade por cabeçalhos — HN 40952304.
- Configuração como código em YAML vista como vantagem — HN 40953205.
- Thread de lançamento com 293 pontos e 92 comentários — HN 26409820 (2021).

## Do que a comunidade se queixa (com evidência)
- **Sem UI e curva de configuração**: "gastei 8h no Authelia; 30 min no authentik" — HN 40953205 (2024).
- **Lacunas de protocolo** nas issues mais reagidas: OIDC Provider ainda aberta como tracking (#189, 119 reações), SAML IdP (#493, 64), OIDC RP/login social (#2728, 61), painel de gestão (#303, 63), recarga de config (#1275, 60).
- OIDC continua "beta" anos depois do início (v4.29) e logout federado ainda não começou.

## Histórico de segurança
Poucos advisories (7), mais leves que a média do grupo:

| ID | Ano | Classe | Lição |
|---|---|---|---|
| CVE-2021-32637 (alto) | 2021 | Bypass de autenticação com URI malformada no nginx | Forward auth depende da URI repassada pelo proxy; canonicalizar e testar cada proxy |
| CVE-2021-29456 (médio) | 2021 | Open redirect no logout | Validar destino de redirect também no logout |
| GHSA-x883-2vmg-xwf7 (baixo) | 2024 | Mudanças de grupo no backend YAML com efeito inesperado | Cache de grupos precisa invalidar sessão/decisão |
| CVE-2025-24806 (baixo) | 2025 | Regulation contada separadamente para login por username e por e-mail | Limites de tentativa por identidade canônica, não por identificador |
| CVE-2026-47203 (médio) | 2026 | Falta de canonicalização de username no Basic Auth com LDAP | Canonicalizar identificador antes de autenticar e de aplicar política |
| CVE-2026-48794 (baixo) | 2026 | Regra de domínio não casada por falta de canonicalização | Canonicalização de host é parte do motor de política |
| CVE-2026-33525 (baixo) | 2026 | XSS potencial | — |

Padrão: **canonicalização** (URI, host, username) é a classe recorrente.

## Arquitetura e decisões notáveis
- Fica ao lado do proxy, nunca no caminho do payload (só vê cabeçalhos de autenticação); APIs protegidas podem ser REST, GraphQL etc.
- Só HTTPS: HTTP recusado de propósito para evitar erro de configuração (#590).
- Stateless opcional: memória/SQLite/arquivo são aceitos para uso simples, mas a doc classifica cada um por severidade para HA.
- Passkey sozinha conta como 1 fator por padrão; opção experimental para aceitar passkey com user verification como 2FA (decisão conservadora explícita).
- Configuração OIDC migrará de YAML para "in-storage" (rotação de JWK, DCR) na v4.40.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Multi-tenancy: não; "multi-issuer" por domínio planejado como breaking change.
- Extensibilidade: baixa (sem hooks/plugins); integração por cabeçalhos e OIDC.
- Configuração: YAML + variáveis de ambiente + segredos em arquivo; sem UI.
- Bancos: SQLite, MySQL/MariaDB, PostgreSQL. Soberania total (self-hosted, sem nuvem).

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **não**.
- **Autorização sensível ao estado / transições**: **não**. Regras de acesso por requisição HTTP (domínio, recurso, método), sem estado de recurso.

## Implicações para o projeto
- **Adotar**: modo proxy/forward auth com regras declarativas por domínio/caminho/método e nível de garantia (AMR/ACR), cobrindo o modelo "proxy" da premissa 1.
- **Adotar**: canonicalização central de identificadores e URIs como requisito de segurança testado em CI.
- **Adotar**: tabela de "severidade por componente stateful" na doc para orientar HA (premissa 17).
- **Adotar**: agilidade criptográfica que permita ML-DSA quando bibliotecas padrão suportarem (sem cripto própria).
- **Evitar**: configuração só em arquivo sem UI administrativa (queixa principal; contraria premissa 20).
- **Oportunidade**: público homelab/PME quer proxy simples + IdP completo com UI; nenhum dos dois (Authelia, authentik) satisfaz leveza e completude ao mesmo tempo.

## Fontes
- https://github.com/authelia/authelia — 2026-11-03
- https://api.github.com/repos/authelia/authelia — 2026-11-03
- https://api.github.com/repos/authelia/authelia/security-advisories — 2026-11-03
- https://api.github.com/search/issues?q=repo:authelia/authelia+is:issue&sort=reactions — 2026-11-03
- https://www.authelia.com/overview/prologue/introduction/ — 2026-11-03
- https://www.authelia.com/overview/prologue/architecture/ — 2026-11-03
- https://www.authelia.com/overview/authentication/security-key/ — 2026-11-03
- https://www.authelia.com/overview/authorization/statelessness/ — 2026-11-03
- https://www.authelia.com/overview/security/post-quantum-cryptography/ — 2026-11-03
- https://www.authelia.com/configuration/storage/introduction/ — 2026-11-03
- https://www.authelia.com/configuration/security/access-control/ — 2026-11-03
- https://www.authelia.com/configuration/identity-validation/elevated-session/ — 2026-11-03
- https://www.authelia.com/integration/openid-connect/introduction/ — 2026-11-03
- https://www.authelia.com/roadmap/ — 2026-11-03
- https://www.authelia.com/roadmap/active/openid-connect-1.0-provider/ — 2026-11-03
- https://www.authelia.com/roadmap/active/granular-authorization/ — 2026-11-03
- https://news.ycombinator.com/item?id=26409820 , 40951166 (comentários 40952107, 40952304, 40953205), 44018802 — 2026-11-03
