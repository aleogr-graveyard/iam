# Autenticação em frameworks tradicionais (Django Auth, Devise, ASP.NET Core Identity, Spring Security + Authorization Server)

| Campo | Valor |
|---|---|
| Categoria | Módulos de autenticação embutidos em frameworks web (biblioteca dentro do monólito); referência de "mínimo esperado" e de longevidade |
| Site / repositório | Django `django.contrib.auth` (https://github.com/django/django) · Devise (https://github.com/heartcombo/devise) · ASP.NET Core Identity (https://github.com/dotnet/aspnetcore, `src/Identity`) · Spring Security, que absorveu o Spring Authorization Server na 7.0 (https://github.com/spring-projects/spring-security) |
| Licença e modelo de negócio | Todos abertos e gratuitos: Django BSD-3-Clause; Devise MIT; ASP.NET Core MIT; Spring Security Apache-2.0. Monetização indireta (fundações, Microsoft, Broadcom/VMware Tanzu com suporte comercial). Exceção relevante: o servidor OIDC "padrão" do ecossistema .NET, **Duende IdentityServer, exige licença paga** em produção (a própria doc da Microsoft avisa) |
| Stack | Python/ORM Django · Ruby on Rails/Warden/ActiveRecord · C#/EF Core · Java/Kotlin (Servlet e WebFlux), OpenSAML 5, Nimbus JOSE |
| Maturidade | Django ~91,5 mil estrelas, Scorecard 6,7 · Devise ~24,4 mil estrelas, Scorecard 4,2 (Maintained 3/10), 5.0.0 em 23/01/2026, 5.0.4 em 05/2026 · aspnetcore ~38,5 mil estrelas, 4.317 issues abertas · Spring Security ~9,7 mil estrelas, Scorecard 6,8 (Fuzzing 10, SAST 10), 7.2.0-RC1 em 22/10/2026; Spring Authorization Server ~5,1 mil estrelas (deps.dev, Scorecard de 24/08/2026) |
| Acesso | 2026-11-03 |

> Limitação: `docs.djangoproject.com`, `learn.microsoft.com`, `guides.rubyonrails.org` não estavam liberados na rede da sessão. Li código e documentação nos repositórios (raw.githubusercontent.com), advisories via OSV e buscas. Itens "(conhecimento prévio)" não foram reconferidos.

## Resumo
- São **kits de construção**: senha, sessão, recuperação e (às vezes) TOTP; o resto vem de pacotes da comunidade (allauth, devise-two-factor, OmniAuth, OpenIddict).
- Spring Security é o mais completo hoje: MFA declarativo, passkeys, one-time token, X.509, LDAP, SAML2 e, desde a 7.0, um Authorization Server com OAuth 2.1/OIDC, PAR, DPoP, device flow, token exchange e mTLS.
- ASP.NET Core Identity ganhou passkeys no .NET 10, mas deixa OIDC/OAuth server para terceiros (Entra, OpenIddict, Duende pago).
- Django e Devise param no básico; Rails 8 trouxe um gerador de autenticação próprio que compete com o Devise.
- Histórico de CVEs longo e instrutivo: enumeração por tempo, normalização Unicode, truncamento do bcrypt, *bypass* de autorização por casamento de caminhos.

## Funcionalidades por domínio (5.1)

Legenda: D = Django (núcleo), V = Devise (núcleo), A = ASP.NET Core Identity, S = Spring Security 7.x (com Authorization Server). Entre parênteses, o que exige pacote de terceiros.

| # | Domínio | D | V | A | S | Notas |
|---|---|---|---|---|---|---|
| 1 | Autenticação primária | ◐ | ◐ | ◐ | ● | Senha em todos. Social/OIDC: D (allauth), V (OmniAuth), A (handlers OIDC/Google/Microsoft da Microsoft), S (OAuth2 Login nativo). SAML: D (allauth/python3-saml), V (omniauth-saml), A (terceiros), S (SAML2 SP nativo). Passkeys: D (allauth, desativado por padrão), A (.NET 10, nativo), S (nativo). LDAP: S nativo; demais por pacote. Link mágico: S (one-time token). ICP-Brasil: nenhum; S tem X.509 (mTLS) que serviria de base |
| 2 | MFA e step-up | ○ | ○ | ◐ | ◐ | D: allauth MFA (TOTP, WebAuthn, códigos de recuperação). V: devise-two-factor. A: TOTP + códigos de recuperação nativos; passkey **não** é segundo fator. S 7.0: MFA por "factor authorities" (ex.: exigir `FACTOR_PASSWORD` nos últimos 30 min = step-up). Push/adaptativo: nenhum |
| 3 | Ciclo de vida da conta | ◐ | ◐ | ◐ | ○ | D: usuário, reset sem estado, ativo/inativo. V: módulos Registerable, Confirmable, Recoverable, Lockable. A: confirmação, reset, lockout, external logins. S: não gerencia usuários (só `UserDetailsService`). SCIM, convites, verificação de idade: nenhum |
| 4 | Sessões e tokens | ◐ | ◐ | ◐ | ● | D/V/A: sessão por cookie do framework, rotação de chave de sessão no login; listagem por dispositivo: não nativo. A: Identity API endpoints com bearer tokens (conhecimento prévio). S: sessão + controle de sessões concorrentes, OIDC back-channel logout (cliente), DPoP no resource server e no AS, refresh tokens |
| 5 | Protocolos | ○ | ○ | ◐ | ● | S AS: OAuth 2.1 (rascunho), OIDC Core/Discovery/Logout/DCR, PKCE (padrão na 7.0), PAR (RFC 9126), DPoP (RFC 9449), device flow (RFC 8628), token exchange (RFC 8693), mTLS (RFC 8705), `private_key_jwt`. A: só cliente OIDC; servidor via OpenIddict/Duende. CIBA, FAPI: não encontrados em nenhum |
| 6 | Máquinas e agentes | ○ | ○ | ○ | ◐ | S AS: client credentials, token exchange, mTLS. Chaves de API, workloads, agentes: nenhum nativo |
| 7 | Detecção e defesa | ○ | ◐ | ◐ | ○ | V: Lockable. A: lockout nativo. D: nada (django-axes). Rate limiting genérico: ASP.NET Core tem middleware de rate limiting (conhecimento prévio). Senhas vazadas/bots: nenhum |
| 8 | Privacidade e conformidade | ○ | ○ | ◐ | ○ | A: `PersonalData` (download/exclusão de dados pessoais no template) (conhecimento prévio). Demais: responsabilidade da aplicação |
| 9 | Experiência e administração | ◐ | ◐ | ◐ | ○ | Telas geradas (D admin + views; V views geráveis; A Identity UI/Blazor). i18n pelo framework. Webhooks/eventos: sinais do Django; nada padronizado |
| 10 | Operação e migração | ◐ | ◐ | ◐ | ◐ | Rehash transparente de senha no login (D: troca de hasher/iterações; A: `PasswordHasherCompatibilityMode`; S: `DelegatingPasswordEncoder` com prefixo `{id}`) |
| 11 | Identidade descentralizada | ○ | ○ | ○ | ○ | Nenhum |

## Diferenciais / exclusivos
- **Spring Security: MFA como autorização** — cada método adiciona uma *authority* de fator com carimbo de tempo; regras declaram quais fatores (e quão recentes) cada rota exige. Elegante para step-up configurável.
- **Spring `DelegatingPasswordEncoder`**: hash armazenado com prefixo do algoritmo (`{bcrypt}`, `{argon2}`), migração gradual sem flag de versão.
- **Django**: PBKDF2-SHA256 com **1.800.000 iterações** no `main` e aumento automático a cada versão; `PASSWORD_HASHERS` ordenado permite migrar algoritmo no login (Argon2, bcrypt-SHA256, scrypt disponíveis).
- **ASP.NET Core Identity passkeys**: deliberadamente restrito a autenticação; sem validação de atestação por padrão (gancho para customizar); exige esquema `Version3`.
- **Devise**: modularidade por *concerns* (10 módulos ligáveis no modelo) — inspiração para composabilidade, com o custo de acoplamento ao modelo `User`.

## O que a comunidade elogia (com evidência)
- Longevidade e "chato no bom sentido": comentários no HN contrapõem frameworks maduros ao *churn* do ecossistema JS ("in rails you can use the rails 8 auth... before it was devise. java - spring security" — [HN 45389293](https://news.ycombinator.com/item?id=45389293)).
- Rails 8 incluiu gerador de autenticação ([notas da 8.0](https://github.com/rails/rails/blob/main/guides/source/8_0_release_notes.md), PR #52328), sinal de que o "mínimo esperado" cabe no framework.
- Spring Security: integração profunda com o ecossistema, processo de segurança maduro (Scorecard Fuzzing 10/SAST 10).
- django-allauth cobre social, MFA, WebAuthn e modo *headless* em um pacote só ([notas de versão](https://docs.allauth.org/en/dev/release-notes/recent.html)).

## Do que a comunidade se queixa (com evidência)
- Spring Security: **complexidade** e configuração difícil ("just complex things" — [HN 45389293](https://news.ycombinator.com/item?id=45389293)); 1.556 issues abertas (deps.dev).
- Devise: manutenção lenta (Scorecard Maintained 3/10; sem política de segurança no Scorecard) e "mágica" difícil de customizar; migração para o gerador do Rails 8 ("authentication-zero... no need for 3rd party provider" — [HN 48038827](https://news.ycombinator.com/item?id=48038827)).
- .NET: o servidor OIDC "oficial" dos templates virou **produto pago** (Duende); a doc da Microsoft avisa sobre licença ([AspNetCore.Docs](https://github.com/dotnet/AspNetCore.Docs/blob/main/aspnetcore/includes/DuendeIdentityServer.md)).
- Django: nada de MFA/rate limit no núcleo; tudo depende de pacotes de terceiros com qualidade variável (conhecimento prévio).

## Histórico de segurança
Fonte principal: OSV (PyPI `django`, RubyGems `devise`/`devise-two-factor`/`doorkeeper`, NuGet `Microsoft.AspNetCore.Identity`, Maven `org.springframework.security:*`).

| ID | Ano | Produto / classe | Lição para nós |
|---|---|---|---|
| CVE-2019-19844 | 2019 | Django — e-mail com Unicode que, após *case folding*, casa com outro usuário recebe o token de reset (crítico) | Normalizar e comparar identificadores de forma canônica; enviar sempre ao e-mail armazenado |
| CVE-2024-39329, CVE-2024-45231 | 2024 | Django — enumeração de usuários por tempo (senha inutilizável) e de e-mails no reset | Tempo constante e respostas idênticas em login/reset |
| CVE-2025-22228 → CVE-2025-22234 | 2025–26 | Spring — `BCryptPasswordEncoder` aceitava senhas > 72 bytes com mesmo prefixo; a correção **quebrou a mitigação de timing** | Limitar entrada antes do hash; testes de regressão de propriedades de segurança (timing) |
| CVE-2026-22746 | 2026 | Spring — enumeração de atributos de usuário em `DaoAuthenticationProvider` | Idem |
| CVE-2024-22258 | 2024 | Spring AS — *downgrade* de PKCE para clientes confidenciais | Uma vez iniciado com PKCE, exigir PKCE até o fim |
| CVE-2026-41008 | 2026 | Spring AS — `request_uri` (PAR) mal validado → open redirect | Validar `redirect_uri` antes de qualquer erro redirecionável |
| CVE-2026-41694, -40988, -40993, -41003 | 2026 | Spring SAML2 — oráculo de decifração sem assinatura, bomba de descompressão, desserialização, execução em formulários | Assinatura antes de decifrar; limites de tamanho; nada de desserialização genérica |
| CVE-2026-22747, CVE-2026-47838 | 2026 | Spring — CN malformado em certificado X.509 permite impersonar outro usuário | **Direto para ICP-Brasil**: extrair identidade (CPF/CNPJ) de campos específicos com parser estrito, nunca de CN textual |
| CVE-2023-34034/34035, CVE-2024-38821, CVE-2026-22753/22754 | 2023–26 | Spring — *bypass* de autorização por casamento de caminhos/servlets | Autorizar sobre representação canônica única da rota |
| CVE-2025-22223, -41232, -41248 | 2025 | Spring — anotações de segurança ignoradas (herança/genéricos/privados) | Negar por padrão; testes que provam que rotas sem regra falham |
| CVE-2023-33170 | 2023 | ASP.NET Identity — lockout não atualizado imediatamente (corrida) | Contadores de falha atômicos |
| CVE-2025-24070 | 2025 | ASP.NET Identity — `RefreshSignInAsync` com usuário errado → entrar como outro (EoP) | APIs internas não devem aceitar "qual usuário" do chamador quando há sessão |
| CVE-2019-16109, CVE-2026-32700 | 2019/26 | Devise — token de confirmação em branco aceito; corrida no "trocar e-mail" confirma e-mail alheio | Tokens nunca vazios; troca de e-mail atômica e vinculada ao valor pedido |
| CVE-2015-7225, CVE-2021-43177, CVE-2024-8796 | 2015–24 | devise-two-factor — OTP não "queimado" após uso; segredo TOTP curto por padrão | TOTP: impedir reuso na janela; segredo ≥ 160 bits |

## Arquitetura e decisões notáveis
- Todos são **in-process** e acoplados ao modelo de usuário/ORM do framework; nenhum separa AuthN como serviço.
- Abstrações de extensão: Django *authentication backends*; Warden *strategies*; ASP.NET `IUserStore`/`SignInManager`; Spring `AuthenticationProvider`/`SecurityFilterChain`/`AuthorizationManager` (o mais próximo de "strategy + middleware" da 4.30).
- Spring incorporou o Authorization Server ao projeto principal (mesmas coordenadas Maven; anúncio de 11/09/2025): sinal de que OAuth/OIDC server virou parte do "núcleo".

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Multi-tenancy: nenhum nativo (Spring tem `AuthenticationManagerResolver` por issuer para resource server multi-tenant — conhecimento prévio).
- Configuração: em código (settings.py, initializers, `Program.cs`, beans); nada por tenant em runtime.
- Bancos: qualquer um suportado pelo ORM; Spring é agnóstico (o usuário traz o store).
- Soberania: total (dados da aplicação).

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **não** em nenhum dos quatro.
- **Autorização sensível ao estado / transições declarativas**: **não** nativo. Spring permite `AuthorizationManager` com acesso ao objeto (ex.: `@PreAuthorize` com SpEL sobre o recurso), o que torna possível regra por estado escrita à mão — **parcial**, sem modelo declarativo de transições. Django tem permissões por objeto apenas como API para backends de terceiros (conhecimento prévio).

## Implicações para o projeto
1. **Adotar** o modelo de MFA do Spring 7 (fatores como atributos com tempo; políticas declaram fatores exigidos e frescor) — encaixa na premissa 3 (composabilidade por configuração).
2. **Adotar** hash com identificador de algoritmo/parâmetros e *rehash* no login (Django/Spring); default Argon2id, parâmetros configuráveis com alerta.
3. **Adotar** checklist de regressão das classes clássicas: enumeração por tempo, Unicode em e-mail, limite de 72 bytes (se bcrypt), reuso de TOTP, segredo TOTP curto, tokens vazios.
4. **Evitar** autorização baseada em casamento textual de caminhos; rotas canônicas e *deny by default* com teste automático.
5. **ICP-Brasil**: tratar certificados com parser estrito de OIDs/SAN (lição CVE-2026-22747), não por CN.
6. **Oportunidade**: nenhum desses oferece multi-tenancy, configuração em runtime, SCIM, detecção de ataques ou M de N; quem sai deles para escalar procura exatamente isso.
7. **Referência de escopo**: o conjunto de RFCs do Spring Authorization Server é um bom "mínimo de protocolo" para o nosso OP (PAR, DPoP, device, token exchange, mTLS, DCR, PKCE obrigatório).

## Fontes
- https://raw.githubusercontent.com/django/django/main/django/conf/global_settings.py e `django/contrib/auth/hashers.py` — 2026-11-03
- https://raw.githubusercontent.com/heartcombo/devise/main/README.md — 2026-11-03
- https://rubygems.org/gems/devise (via busca: 5.0.0 em 23/01/2026; 5.0.4 em 08/05/2026) — 2026-11-03
- https://github.com/rails/rails/blob/main/guides/source/8_0_release_notes.md — 2026-11-03
- https://raw.githubusercontent.com/dotnet/aspnetcore/main/src/Identity/Extensions.Core/src/PasswordHasherOptions.cs — 2026-11-03
- https://github.com/dotnet/AspNetCore.Docs/blob/main/aspnetcore/includes/DuendeIdentityServer.md e `aspnetcore/security/authentication/passkeys/index.md` — 2026-11-03 — documentação oficial
- https://andrewlock.net/exploring-dotnet-10-preview-features-6-passkey-support-for-aspnetcore-identity (via busca) — 2026-11-03 — blog independente
- https://raw.githubusercontent.com/spring-projects/spring-security/main/docs/modules/ROOT/nav.adoc, `.../pages/servlet/oauth2/authorization-server/index.adoc`, `7.0.x/.../whats-new.adoc` — 2026-11-03
- https://docs.spring.io/spring-security/reference/7.0/servlet/authentication/mfa.html (via busca) — 2026-11-03
- https://repo1.maven.org/maven2/org/springframework/security/spring-security-core/maven-metadata.xml — 2026-11-03
- https://thenote.app/post/en/spring-authorization-server-moving-to-spring-security-7-0-gsm2ta4qir (via busca; anúncio de 11/09/2025) — 2026-11-03
- https://docs.allauth.org/en/dev/release-notes/recent.html (via busca) — 2026-11-03
- https://api.osv.dev (pacotes citados) — 2026-11-03
- https://api.deps.dev/v3/projects/ (django/django, heartcombo/devise, dotnet/aspnetcore, spring-projects/spring-security, spring-projects/spring-authorization-server) — 2026-11-03
- https://news.ycombinator.com/item?id=45389293, https://news.ycombinator.com/item?id=48038827 — 2026-11-03
