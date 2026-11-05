# oauth2-proxy

| Campo | Valor |
|---|---|
| Categoria | AuthN, proxy de autenticação (reverse proxy ou middleware `auth_request`/forward-auth) diante de aplicações; delega a identidade a um IdP OAuth2/OIDC |
| Site / repositório | https://oauth2-proxy.github.io/oauth2-proxy/ · https://github.com/oauth2-proxy/oauth2-proxy |
| Licença e modelo de negócio | MIT, 100% comunitário, sem edição paga nem empresa por trás. Patrocínio do SAP Open Source Program (antes créditos Azure da Microsoft). CNCF Sandbox desde 02/out/2025; site sob "a Series of LF Projects, LLC". Sem mudança de licença |
| Stack | Go; binário único, imagem distroless desde v7.6.0 (variante `-alpine`); sessões em cookie (padrão) ou Redis. Sem banco |
| Maturidade | Origem no `bitly/oauth2_proxy` (citado no HN já em 2017/2018), hoje na org `oauth2-proxy`. ~15,3 mil estrelas, ~2,3 mil forks, 159 issues abertas. Versão 7.15.5 (01/out/2026); cadência de patch mensal a bimestral. Adoção ampla em Kubernetes (ingress-nginx, Traefik), charm Canonical, apps no Unraid |
| Acesso | 2026-11-03 |

> Limitações: `api.github.com` bloqueado na sessão (sem métricas via API); `oauth2-proxy.github.io` lido via curl. Orçamento de buscas web esgotado no fim da pesquisa.

## Resumo
- Proxy leve e sem estado que coloca login OIDC/OAuth2 na frente de qualquer app HTTP e repassa identidade em cabeçalhos.
- Não é IdP: não guarda usuários, senhas nem fatores; todo o "AuthN real" é do provedor (Google, Entra ID, GitHub, Keycloak, OIDC genérico...).
- Autorização grossa: domínio de e-mail, lista de e-mails, grupos/organizações por provedor.
- Projeto voluntário; demandas antigas (múltiplos provedores, back-channel logout, logs JSON) seguem abertas há 4–6 anos.
- Em 2025–2026 acumulou vários bypasses de autenticação críticos ligados a interpretação de caminho e de cabeçalhos `X-Forwarded-*`.

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ◐ | Só login federado: 19 provedores (ADFS, Bitbucket, Cidaas, Cisco Duo, DigitalOcean, Facebook, Gitea/Forgejo, GitHub, GitLab, Google, Keycloak OIDC, LinkedIn, login.gov, Entra ID, Nextcloud, SourceHut, OIDC genérico...). Senha, sem senha, passkeys, SAML, LDAP: delegados ao IdP. "Multiple providers" incompleto (doc de provedores; issue #926 aberta desde 2020, 86 reações) |
| 2 | MFA e step-up | ○ | Delegado ao IdP. Há apenas a opção `acr_values` para pedir nível ao IdP; sem step-up por rota |
| 3 | Ciclo de vida da conta | ○ | Inexistente (não guarda contas). `--authenticated-emails-file` como lista de permissão |
| 4 | Sessões e tokens | ◐ | Sessão em cookie cifrado (stateless) ou Redis com "ticket" (ID 128 bits + segredo por sessão; valor cifrado). Refresh de tokens do IdP; sessão invalidada em erro fatal de refresh (v7.15.2). Logout só local (`/oauth2/sign_out`, com `rd` para o IdP); back-channel logout pedido desde 2021 (#1224, aberta, 54 reações). DPoP: não |
| 5 | Protocolos | ◐ | Cliente (RP) OAuth2/OIDC; algoritmos de assinatura JWT configuráveis (v7.15.0). Não é servidor OAuth/OIDC/SAML; sem SCIM |
| 6 | Máquinas e agentes | ◐ | Aceita bearer JWT de emissores extras (`skip_jwt_bearer_tokens`/`extra_jwt_issuers`, opção de config) e injeta claims; sem contas de serviço próprias nem chaves de API |
| 7 | Detecção e defesa | ○ | Sem rate limiting, proteção contra bots ou senhas vazadas (fica para o IdP e o proxy de borda) |
| 8 | Privacidade e conformidade | ○ | Não trata dados pessoais além da sessão; sem recursos de conformidade |
| 9 | Experiência e administração | ◐ | Página de login simples customizável por templates; sem console administrativo; configuração só por arquivo/flags |
| 10 | Operação e migração | ◐ | Binários multi-arquitetura (inclui ppc64le, s390x), distroless, `/ping`, `/ready`, `/metrics` Prometheus, `--config-test` para CI (v7.15.0). Logs estruturados JSON ainda pedidos (#1057, aberta, 50 reações) |
| 11 | Identidade descentralizada | ○ | Não |

## Diferenciais / exclusivos
- **Simplicidade e ausência de estado**: modo cookie dispensa qualquer armazenamento; escala horizontal trivial.
- **Dois modos de implantação**: reverse proxy completo ou só endpoint `/oauth2/auth` (202/401) para `auth_request` do nginx / forward-auth do Traefik.
- **Provedores "especializados"** extraem grupos/organizações (GitHub org/team, GitLab group, Google groups, Entra ID) e repassam em cabeçalhos.
- **Desenho de sessão em Redis**: o cookie leva um ticket com segredo por sessão, então o vazamento do Redis sozinho não revela sessões (doc de session storage).

## O que a comunidade elogia (com evidência)
- Leve e "boring": recomendado como alternativa open source ao Cloudflare Access e ao Google IAP desde 2017 ([HN 16167964](https://news.ycombinator.com/item?id=16167964), [HN 13833855](https://news.ycombinator.com/item?id=13833855)).
- Comparativos de 2026 o apontam como a opção mais leve para homelab/autohospedagem ([pistack.xyz](https://www.pistack.xyz/posts/oauth2-proxy-vs-pomerium-vs-traefik-forward-auth-2026/) — blog independente, sem verificação de metodologia).
- Projeto com boas práticas visíveis: OpenSSF Scorecard e Best Practices, FOSSA, cobertura medida (badges no README).

## Do que a comunidade se queixa (com evidência)
Issues mais votadas ([GitHub, por reações](https://github.com/oauth2-proxy/oauth2-proxy/issues?q=is%3Aissue+sort%3Areactions-%2B1-desc)):
- **Um provedor por instância**: #926 "Enable multiple providers" (aberta desde nov/2020, 86).
- **Regressões em upgrades**: #1666 "v7.3.0 breaks azure provider" (71), #1724 "github broken in 7.3.0" (28); mudanças "breaking" no Alpha Config (v7.14.0 exige fontes de cabeçalho aninhadas).
- **Back-channel logout** (#1224, 54) e **logs JSON** (#1057, 50) pedidos há 5 anos.
- **Exigência de e-mail no OIDC** (#2667, aberta).
- **Ritmo de projeto voluntário**: #2051 "New Release?" (2023); README admite que "review times can vary"; #3208 documentação fora do ar (set/2025).
- **Configuração em transição**: o formato YAML "alpha" convive com flags/TOML há anos e pode quebrar "with no notice" (doc Alpha Configuration).
- Viés: Pomerium descreve "NGINX + OAuth2-Proxy" como difícil de manter ([pomerium.com/comparisons](https://www.pomerium.com/comparisons/nginx-oauth2proxy) — concorrente).

## Histórico de segurança
15 advisories no GitHub ([lista](https://github.com/oauth2-proxy/oauth2-proxy/security/advisories)). Destaques:

| ID | Ano | Severidade | Classe |
|---|---|---|---|
| GHSA-63jm-59jj-478j | 2026 | Crítica | Bypass de autenticação: interpretação inconsistente de caminho em skip-auth |
| GHSA-wr5q-7wxw-x568 | 2026 | Crítica | Bypass via cabeçalhos de IP de cliente forjados |
| GHSA-7x63-xv5r-3p2x (CVE-2026-40575, via agregador) | 2026 | Crítica | Bypass via `X-Forwarded-Uri` repassado do cliente (nginx) |
| GHSA-5hvv-m4w4-gf6v (CVE-2026-34457, via agregador) | 2026 | Crítica | Bypass: match de User-Agent de health check vale no modo `auth_request` |
| GHSA-pxq7-h93f-9jrg | 2026 | Alta | Confusão de fragmento (`#`) em regex de rotas skip-auth |
| GHSA-c5c4-8r6x-56w3 | 2026 | Moderada | Bypass de domínio de e-mail com claim malformada com vários `@` |
| GHSA-hhqp-vx7f-5c6m | 2026 | Moderada | Credenciais em log de erro no callback OAuth |
| GHSA-vjrc-mh2v-45x6 | 2025 | Alta | Smuggling de cabeçalho via `_` (underscore) → escalada de privilégio |
| GHSA-7rh7-c77v-6434 / CVE-2025-54576 | 2025 | Crítica (CVSS 9.1) | `skip_auth_routes` casava contra a URI completa, inclusive query string (CWE-290) |
| GHSA-4mf2…, GHSA-5m6c…, GHSA-j7px…, GHSA-qqxw… | 2020–2021 | Alta/Moderada | Série de open redirects (espaço codificado, `/\dominio`, subdomínios) |
| GHSA-652x-m2gr-hppm | 2021 | Moderada | Regressão: flag de grupo GitLab deixou de autorizar na v7.0.0 |

Lições para nós:
- **Diferenças de parsing (path, query, fragmento, encoding) entre proxy e upstream são a principal classe de falha** de proxies de auth. Normalizar uma vez, casar sobre a forma canônica, testar com fuzzing diferencial.
- **Nunca confiar em `X-Forwarded-*`/IP do cliente por padrão**: só de proxies declarados (o projeto criou `--trusted-proxy-ip` em 2026, tarde). Padrão seguro desde o início.
- **Listas de exceção ("skip auth", health checks por User-Agent) são superfície de bypass**: preferir rotas explícitas e separadas, com alerta de configuração (premissa 4).
- **Validar e-mail/claims com parser estrito** (RFC 5322) antes de autorizar por domínio.
- **Open redirect** é recorrente em qualquer `rd`/`redirect_uri`: lista de permissão exata.
- **Regressões funcionais viram falhas de autorização** (GitLab group): testes de regressão para todo critério de autorização.

## Arquitetura e decisões notáveis
- Endpoints fixos: `/oauth2/start`, `/oauth2/callback`, `/oauth2/sign_in`, `/oauth2/sign_out`, `/oauth2/userinfo`, `/oauth2/auth` (202/401), `/ping`, `/ready`, `/metrics` (doc Endpoints).
- Cookie: assinatura e cifragem com `cookie-secret` obrigatório (AES); no modo cookie não há lock de sessão, então refresh concorrente pode forçar novo login (doc Session Storage).
- Repasse de identidade por cabeçalhos (`X-Forwarded-User`, `X-Forwarded-Email`, `Authorization: Bearer` com ID token); Alpha Config permite injetar claims e segredos em cabeçalhos.
- Precedência de configuração: flags > variáveis de ambiente > arquivo (doc Overview).
- Desde 2023 imagens nightly de `master` (quay.io).

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- **Multi-tenancy**: nenhuma; uma instância por provedor/conjunto de regras.
- **Extensibilidade**: novos provedores só por código Go no pacote `providers` (doc de provedores); sem plugins.
- **Configuração**: flags, env, TOML; YAML "alpha" estruturado (upstreams, providers, injeção de cabeçalhos).
- **Bancos**: nenhum; Redis opcional (inclui Sentinel/Cluster).
- **Soberania**: totalmente self-hosted; dados pessoais ficam com o IdP escolhido.

## Verificação de diferenciais candidatos do projeto
- Aprovação multipartes (M de N): **não**. Sem fluxo de solicitação/aprovação.
- Autorização sensível ao estado / transições declarativas: **não**. Autoriza só por identidade (e-mail, domínio, grupo) e rota.

## Implicações para o projeto
1. **Adotar** um modo "proxy/forward-auth" (endpoint 2xx/401) do nosso AuthN: é o modelo de uso que o mercado espera para apps legadas.
2. **Adotar** sessão do lado do servidor com ticket + segredo por sessão no cookie (vazamento do armazenamento não basta para sequestro).
3. **Evitar** a classe de bugs de 2025–2026: parser canônico de URL compartilhado, sem confiança em `X-Forwarded-*` por padrão, sem bypass por User-Agent; testes diferenciais proxy×upstream no CI ofensivo (4.19).
4. **Oportunidade**: múltiplos IdPs por instância, back-channel logout e logs JSON são pedidos há 5+ anos; vir de fábrica no nosso proxy.
5. **Adotar** `--config-test` equivalente (validação de configuração para CI) e alertas de configuração fraca, alinhados à premissa 4.
6. **Evitar** formato de configuração "alpha" eterno: um esquema versionado único, com migração automática.

## Fontes
- https://github.com/oauth2-proxy/oauth2-proxy — 2026-11-03
- https://raw.githubusercontent.com/oauth2-proxy/oauth2-proxy/master/README.md — 2026-11-03
- https://github.com/oauth2-proxy/oauth2-proxy/releases — 2026-11-03
- https://github.com/oauth2-proxy/oauth2-proxy/security/advisories (págs. 1–2) — 2026-11-03
- https://github.com/oauth2-proxy/oauth2-proxy/issues?q=is%3Aissue+sort%3Areactions-%2B1-desc — 2026-11-03
- https://oauth2-proxy.github.io/oauth2-proxy/configuration/providers — 2026-11-03
- https://oauth2-proxy.github.io/oauth2-proxy/configuration/session_storage — 2026-11-03
- https://oauth2-proxy.github.io/oauth2-proxy/features/endpoints — 2026-11-03
- https://oauth2-proxy.github.io/oauth2-proxy/configuration/overview — 2026-11-03
- https://oauth2-proxy.github.io/oauth2-proxy/configuration/alpha-config — 2026-11-03
- https://www.cncf.io/projects/oauth2-proxy/ — 2026-11-03 (via resultado de busca)
- https://nvd.nist.gov/vuln/detail/CVE-2025-54576 — 2026-11-03 (via resultado de busca; NVD bloqueado na sessão)
- https://vulnerabilities.etecs.ru/vendors/oauth2_proxy_project?lang=en — 2026-11-03 — agregador; CVE-2026-* não confirmados em fonte primária
- https://news.ycombinator.com/item?id=16167964 , https://news.ycombinator.com/item?id=13833855 (via API Algolia) — 2026-11-03
- https://www.pistack.xyz/posts/oauth2-proxy-vs-pomerium-vs-traefik-forward-auth-2026/ — 2026-11-03 — blog independente
- https://www.pomerium.com/comparisons/nginx-oauth2proxy — 2026-11-03 — **concorrente**
