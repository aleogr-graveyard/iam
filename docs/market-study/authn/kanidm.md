# Kanidm

| Campo | Valor |
|---|---|
| Categoria | AuthN, serviço self-hosted (IDM "tudo em um": diretório, IdP OAuth2/OIDC, LDAP de leitura, RADIUS, PAM/NSS para Linux) |
| Site / repositório | https://kanidm.com · https://github.com/kanidm/kanidm · livro: https://kanidm.github.io/kanidm/stable/ |
| Licença e modelo de negócio | MPL-2.0, projeto comunitário, sem edição paga nem SaaS. Sem mudança de licença encontrada |
| Stack | Rust (servidor `kanidmd`, CLI, `kanidm_unixd`), web UI renderizada no servidor com htmx. Banco embutido (arquivo `kanidm.db`, SQLite) com cache próprio; sem suporte a bancos externos. Replicação própria multi-nó |
| Maturidade | Repositório desde fev/2019. ~5,5 mil estrelas, 371 forks, ~280 issues abertas, push semanal. Série 1.x com release trimestral (1º de fev/mai/ago/nov). Usado sobretudo em homelab e pequenas organizações; operador K8s comunitário (Kaniop) |
| Acesso | 2026-11-03 |

> Limitação: GitHub web e parte da API com rate limit; Reddit inacessível. Usei o livro oficial (versão `print.html`), API do GitHub e HN Algolia.

## Resumo
- Foco em **segurança por padrão e credenciais por dispositivo/serviço**: passkeys preferidas, senha de RADIUS separada, application passwords por app, chaves SSH distribuídas.
- Integra Linux (PAM/NSS, offline), Wi-Fi/VPN (RADIUS, EAP-TLS), apps web (OAuth2/OIDC) e legado (LDAPS somente leitura).
- **Sessões em modo leitura com elevação temporária** ("privilege access mode", como `sudo`).
- Lacunas notáveis: sem SAML, sem login social/federação de entrada (em desenvolvimento), sem device flow, sem e-mail/SMS, UI admin limitada (CLI-first).

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ◐ | Passkeys (preferido), attested passkeys (allowlist por AAGUID/CA via FIDO MDS), senha + TOTP, LDAP bind (para apps legadas), RADIUS (MSCHAPv2, EAP-TLS por certificado). Login social e SAML/OIDC de entrada: ausentes no livro estável (advisory de 2026 cita "OAuth2 trust login provider", sinal de recurso em desenvolvimento). Magic link: não |
| 2 | MFA e step-up | ◐ | TOTP (pedido antes da senha, contra brute force), WebAuthn nível 3. Reautenticação obrigatória para privilégios de escrita (privilege expiry ≤ 3600 s). SMS/e-mail/push e MFA adaptativo: ausentes |
| 3 | Ciclo de vida da conta | ◐ | Contas criadas por admin/delegados; link/QR de reset de credencial para autoatendimento; validade de conta; sync de entrada de LDAP (RFC 4533) e FreeIPA; importação SCIM. Cadastro público, verificação por e-mail, exclusão pelo titular: não encontrados |
| 4 | Sessões e tokens | ◐ | Sessões com expiração por política, revogação, refresh token com expiração configurável (#3040), OIDC back-channel logout (#2412 fechada). DPoP: não encontrado |
| 5 | Protocolos | ◐ | OAuth 2.0/OIDC (PKCE só S256, introspecção, revogação, RFC 9068, token exchange RFC 8693, WebFinger), LDAP v3 só leitura, RADIUS, SCIM (somente importação), WebAuthn L3. SAML, device flow (#1523, a mais votada, aberta), CIBA, PAR, FAPI: ausentes |
| 6 | Máquinas e agentes | ● | Service accounts com gestão delegada (`entry_managed_by`), API tokens JWS somente leitura por padrão (escrita explícita), troca de API token por tokens OIDC (RFC 8693), application passwords, chaves SSH. Tokens OAuth para service accounts ainda pedidos (#3109) |
| 7 | Detecção e defesa | ◐ | Política de credenciais por grupo (mínimo `mfa`/`passkey`/`attested_passkey`), comprimento mínimo de senha, TOTP antes da senha. Lista de senhas proibidas, rate limit e anomalias: não verificados no livro |
| 8 | Privacidade e conformidade | ◐ | Escopos PII controlados por cliente OAuth, alinhamento declarado ao NIST SP 800-63, Argon2id, vinculação TPM. Exportação/exclusão pelo titular: não encontrado |
| 9 | Experiência e administração | ◐ | CLI completa; web UI de usuário e admin em evolução (#996 aberta desde 2022); i18n: não verificado. Webhooks: não encontrado |
| 10 | Operação e migração | ◐ | Backup online agendado, restore, verificação de consistência, reindex, replicação multi-nó (consistência eventual forte), migrações SCIM declarativas. Downgrade impossível; renomear domínio quebra passkeys |
| 11 | Identidade descentralizada | ○ | Não encontrado |

## Diferenciais / exclusivos
- **Privilege access mode**: sessão longa só leitura; escrita exige reautenticação recente e expira em até 1 h.
- **Account policy por grupo** com "credential type minimum" e allowlist de autenticadores atestados; políticas múltiplas combinam pelo componente mais estrito.
- **Credenciais por serviço**: RADIUS, application passwords e chaves SSH separadas da credencial primária; vazamento em um serviço não compromete a conta.
- **Proteções contra escalada no help desk**: membros de `idm_high_privilege` não podem ter credenciais redefinidas por papéis inferiores.
- **Integração nativa com Linux** (PAM/NSS, cache offline) e WebFinger para descoberta OIDC.
- Advisory GHSA-hh34-7jqq-3f73: análise do OAuth com a ferramenta **Oauch**, publicada pelo próprio projeto como advisory (transparência).

## O que a comunidade elogia (com evidência)
- Leve e fácil de entender frente ao Keycloak — HN 45916763 (2025), 47651624 (2026, "falta SAML").
- Padrões seguros e LDAP pronto — HN 40952456 (2024).
- Design "tudo em um" para evitar a fragilidade de microserviços do FreeIPA (autor) — HN 28892045 (2021, viés: autor).
- Thread de lançamento com 180 pontos — HN 28886636.

## Do que a comunidade se queixa (com evidência)
- **Senha de RADIUS separada** inviabilizou adoção em telecom: "não dá para ter uma senha única" — HN 44916487 (2025).
- **Exige usuário técnico** — HN 40952456; CLI-first, UI admin incompleta (#996).
- **Lacunas de protocolo**: device flow (#1523, 41 reações), Kerberos (#1614), trusts entre domínios (#149, aberta desde 2019), usernameless WebAuthn (#1542), domain join (#2495), k8s (#282).
- Escopo crescendo (ACME considerado): "por que todo app quer ser minha raiz de confiança?" — HN 48923077 (2026).

## Histórico de segurança
~10 advisories visíveis, quase todos 2026; poucos CVEs atribuídos.

| ID | Ano | Classe | Lição |
|---|---|---|---|
| GHSA-xxwr-vvr3-2g9f (crítico) | 2026 | Tratamento incorreto de modificação de conjuntos permite escrita arbitrária no banco por usuário autenticado | Controle de acesso por atributo precisa de testes de propriedade em operações de conjunto |
| CVE-2026-46689, GHSA-2pm5-6m23-h692 (alto) | 2026 | Stack overflow sem autenticação no parser de filtro SCIM | Parsers recursivos expostos precisam de limite de profundidade e fuzzing (mesmo em Rust) |
| GHSA-qcrp-p3rq-pffr (alto) | 2026 | Filtro LDAP aninhado anônimo derruba o processo | Idem; DoS pré-autenticação é crítico para IdP |
| GHSA-gpxg-fx2g-qxj2 (médio) | 2026 | HTML armazenado no display name → requisição forjada via htmx | UI renderizada no servidor com htmx também precisa de escape contextual |
| GHSA-84jc-3hj2-hwc7 (médio) | 2026 | Validação de imagem antes da autorização; panic em PNG malformado | Autorizar antes de processar entrada |
| GHSA-53hj-r94p-8c8f (baixo) | 2026 | Comparação não constante de `client_secret` | Comparação de segredos sempre em tempo constante |
| GHSA-x8cg-3c8h-gw55 (baixo) | 2026 | Subject do provedor OAuth2 de confiança não validado | Federação de entrada valida `iss`+`sub` |
| GHSA-j4gj-f54h-56j5 (baixo) | 2026 | Módulo RADIUS em Python loga senhas e segredos | Integrações periféricas entram no modelo de ameaças |

Lição geral: memória segura (Rust) não elimina DoS por recursão nem falhas lógicas de autorização.

## Arquitetura e decisões notáveis
- Binário único com banco embutido para testar ponta a ponta e evitar dependências frágeis (lição do FreeIPA, segundo o autor).
- Domínio e origem são imutáveis na prática: WebAuthn amarrado à origem, renomear invalida passkeys; doc dedica capítulo à escolha do domínio (subdomínio exclusivo, evitar wildcard e domínios de teste aninhados).
- Modelo de dados próprio com schema e controle de acesso internos; LDAP é fachada só de leitura.
- Release trimestral, sem downgrade; ferramentas de sync devem ter a mesma versão do servidor.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Multi-tenancy: não (uma instância = um domínio).
- Extensibilidade: baixa; integração por OAuth2/LDAP/RADIUS/SCIM de entrada e REST.
- Configuração: `server.toml` + CLI; migrações SCIM declarativas para gestão de configuração (e Kaniop para GitOps).
- Banco: somente embutido. Soberania total (self-hosted).

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **não**. Há delegação de gestão (`entry_managed_by`) e proteção de alto privilégio, sem aprovação.
- **Autorização sensível ao estado / transições**: **não**.

## Implicações para o projeto
- **Adotar**: sessões com privilégio elevado temporário após reautenticação (ações administrativas e de credencial).
- **Adotar**: política de credencial por grupo/aplicação com nível mínimo (mfa, passkey, passkey atestada) e combinação pelo mais estrito; casa com "métodos classificados por nível de garantia" (premissa 2).
- **Adotar**: credenciais por serviço/dispositivo (app passwords, tokens só leitura por padrão) para limitar raio de dano.
- **Adotar**: capítulo de "escolha de domínio" na doc e alertas de configuração sobre origem WebAuthn (brief 4.6 e 4.17).
- **Evitar**: banco embutido sem opção externa (contraria premissa 9) e senha separada obrigatória sem alternativa configurável (queixa de telecom).
- **Oportunidade**: público quer leveza do Kanidm com SAML, social e UI; parsers (SCIM filter, LDAP filter) devem ter fuzzing desde o início.

## Fontes
- https://github.com/kanidm/kanidm — 2026-11-03
- https://api.github.com/repos/kanidm/kanidm — 2026-11-03
- https://api.github.com/repos/kanidm/kanidm/security-advisories — 2026-11-03
- https://api.github.com/search/issues?q=repo:kanidm/kanidm+is:issue&sort=reactions — 2026-11-03
- https://kanidm.github.io/kanidm/stable/introduction_to_kanidm.html — 2026-11-03
- https://kanidm.github.io/kanidm/stable/print.html (seções Supported Features, Choosing a Domain Name, Accounts, Account Policy, Service Accounts, RADIUS, Backup and Restore, Release Processes) — 2026-11-03
- https://news.ycombinator.com/item?id=28886636 (comentário 28892045, viés: autor), 45916763, 47651624, 40952456, 44916487, 48923077, 47899943 — 2026-11-03
