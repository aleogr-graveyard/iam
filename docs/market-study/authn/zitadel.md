# Zitadel

| Campo | Valor |
|---|---|
| Categoria | AuthN, serviço self-hosted e SaaS (IAM/CIAM com foco em multi-tenancy B2B) |
| Site / repositório | https://zitadel.com · https://github.com/zitadel/zitadel |
| Licença e modelo de negócio | **AGPL-3.0-only** desde a v3 (mar/2025); antes Apache-2.0. Exceções: `proto/` e `apps/docs/` em Apache-2.0; Login V2 (`apps/login/`), `packages/zitadel-client/` e `packages/zitadel-proto/` em MIT. Contribuições externas precisam ser licenciadas em Apache-2.0 (assimetria de entrada/saída que dispensa CLA formal). Licença comercial para quem não aceita AGPL. Receita: Zitadel Cloud (pay-as-you-go; regiões US, EU, AU, CH) e suporte para self-hosted. Mesma base de código no SaaS e no self-hosted |
| Stack | Go (backend), TypeScript/Next.js (Login V2), Angular (Console). APIs gRPC, connectRPC e REST/JSON. Banco: **somente PostgreSQL ≥ 14** desde a v3 (CockroachDB removido). Histórico: event sourcing + CQRS sobre tabela única de eventos |
| Maturidade | Empresa suíça, projeto desde ~2019/2020. ~15,5 mil estrelas, ~1,4 mil forks, ~4,7 mil commits. Release v4.20.0 em 22/out/2026 (cadência semanal/quinzenal). OpenID Certified (OP). v3 EOL em 31/ago/2026 |
| Acesso | 2026-11-03 |

> Limitação: `zitadel.com` e `dev.to` não resolviam na rede da sessão. Usei GitHub (README, LICENSING.md, releases, advisories, issues, clone esparso de `apps/docs/content`), API Algolia do HN e buscas na web (trechos de páginas oficiais).

## Resumo
- IdP em Go, API-first, com hierarquia de multi-tenancy nativa (Instância → Organização → Projeto) e trilha de auditoria completa por eventos.
- Trocou Apache-2.0 por AGPL-3.0 na v3 (2025) e removeu CockroachDB; mantém APIs/SDKs/Login V2 permissivos.
- **Abandona o event sourcing**: a próxima geração (v5/"nextgen") usa tabelas relacionais como fonte de verdade e mantém eventos apenas para auditoria e integrações.
- Onda grande de advisories de tomada de conta em 2026 (Login V1/V2, IdP externo, passkeys, device code).

## Funcionalidades por domínio (5.1)

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ◐ | Usuário/senha, passkeys (FIDO2), login social e IdPs corporativos com templates (OIDC, SAML, LDAP), domain discovery. Link mágico: não encontrado na doc. Certificado X.509/ICP-Brasil: não encontrado |
| 2 | MFA e step-up | ◐ | TOTP, U2F/WebAuthn, OTP por e-mail e SMS (Twilio ou HTTP), políticas de MFA por organização. Códigos de recuperação, push, WhatsApp e MFA adaptativo: não encontrados na doc |
| 3 | Ciclo de vida da conta | ◐ | Autocadastro com verificação de e-mail/telefone, reset, vinculação de contas (inclusive manual), onboarding B2B, convites, SCIM 2.0 **servidor**, JIT via IdP, bloqueio por tentativas (só por contagem; issue #6502 pede bloqueio temporal). SCIM cliente (saída) ausente (#6601). Mover usuário entre organizações ausente (#8835) |
| 4 | Sessões e tokens | ◐ | Session API (fluxos customizados), revogação, refresh tokens, back-channel logout (guia `back-channel-logout`). DPoP ausente (issue #5402 aberta desde 2023). Lifetimes por organização/app ausentes (#5219) |
| 5 | Protocolos | ◐ | OIDC certificado, OAuth 2.0, SAML 2.0 (IdP), SCIM 2.0, WebAuthn, device authorization, token exchange (RFC 8693), Dynamic Client Registration RFC 7591/7592 (desligado por padrão, pensado para clientes MCP). CIBA, PAR e FAPI: não encontrados na doc |
| 6 | Máquinas e agentes | ◐ | Service accounts com JWT Profile, PAT e client credentials; token exchange e impersonação; DCR para MCP. Identidade de workloads e chaves de API para usuários finais: não encontrados |
| 7 | Detecção e defesa | ◐ | Lockout e políticas de senha (complexidade/idade). CAPTCHA, senhas vazadas e detecção de anomalias: não encontrados; a doc de produção recomenda rate limiting e bot detection externos |
| 8 | Privacidade e conformidade | ◐ | Trilha de auditoria completa por eventos, regiões no Cloud (US/EU/AU/CH), tabela de PII na doc legal. Exportação para titular: não verificado |
| 9 | Experiência e administração | ● | Console, Login V2 hospedado ou próprio (Next.js, MIT), branding por organização, i18n, Actions v2 (webhooks, token enrichment), streaming de eventos para SIEM, APIs completas |
| 10 | Operação e migração | ◐ | Atualização sem downtime, escala horizontal sem session store externo, comando `mirror` (CockroachDB → Postgres), importação de usuários. Ciclo de suporte curto (v3 sem correções após 31/ago/2026) |
| 11 | Identidade descentralizada | ○ | "verifiable credential" ausente da doc |

## Diferenciais / exclusivos
- **Multi-tenancy em camadas**: instâncias (tenants de infraestrutura, várias por deployment, roteadas por domínio) + organizações B2B ilimitadas + projetos, com concessão de projetos a outras organizações ("granted projects") e administração delegada.
- **Trilha de auditoria por eventos** consultável por API e exportável via webhooks.
- **API-first real**: tudo em gRPC/connectRPC/REST, incluindo Session API para construir login próprio.
- **Paridade SaaS/self-hosted**: mesmo código.

## Event sourcing: decisão e reversão (ponto central para nós)
- **Desenho original**: eventos imutáveis em tabela única (fonte de verdade), agregados como fronteira de transação, projeções para leitura (CQRS). Benefícios declarados: auditoria ilimitada, "viagem no tempo", recomputar projeções ([doc eventstore](https://github.com/zitadel/zitadel/blob/main/apps/docs/content/concepts/eventstore/overview.mdx)).
- **Problemas relatados pela própria equipe** ([issue #9599](https://github.com/zitadel/zitadel/issues/9599), mar/2025, milestone "Zitadel Version 5", ~47/66 sub-issues concluídas):
  - consultas por campo exigem varredura sequencial; tabela auxiliar "fields" virou "unmaintainable situation";
  - ordenação de eventos força trade-off entre locks e retries;
  - **consistência eventual**: um GET logo após um update pode não refletir a mudança;
  - projeções aplicadas em ordem limitam throughput; projeções travadas/deadlocks recorrentes;
  - sem constraints relacionais: integridade imposta em código.
- Issue #9423 relata uso alto de CPU no PostgreSQL; benchmarks v4 citam over-fetching e consultas a eventos entre as mais caras (via busca).
- **Nova direção** (post "Why Zitadel's Next Generation Is No Longer Event-Sourced", 17/set/2026, via busca): tabelas relacionais atualizadas na mesma transação da requisição são a única fonte de verdade; "wide events" mantidos para auditoria, segurança e integrações; read-your-writes sem esperar projeção; armazenamento relacional para PostgreSQL, Spanner e SQLite (não confirmado no repositório; o README atual já diz "Relational core, event-driven soul").
- **Lição**: auditoria completa não exige event sourcing como modelo de persistência; outbox/evento de auditoria na mesma transação entrega o benefício sem consistência eventual em um IdP.

## Mudança de licença para AGPL
- Anunciada em 2025 com a v3 (prevista para 31/mar/2025); motivo declarado: usuários comerciais de Apache-2.0 contribuíam pouco ([blog](https://zitadel.com/blog/apache-to-agpl), via busca; viés: fornecedor).
- Não retroativa: código e forks anteriores seguem Apache-2.0; Cloud não afetado; licença comercial disponível ([FAQ](https://zitadel.com/license-faq), via busca).
- Desenho prático: núcleo AGPL; protos, SDKs e Login V2 permissivos para não "contaminar" integrações; entrada de contribuições em Apache-2.0.
- Para nós: **não estudar nem reaproveitar código AGPL**; estudar só decisões publicadas.

## O que a comunidade elogia (com evidência)
- Desenho limpo e moderno frente ao Keycloak; Go, binário único ([HN 31408059](https://hn.algolia.com/api/v1/items/31408059), `IceDane`; [HN 31258469](https://hn.algolia.com/api/v1/items/31258469), comentário de autor do Zitadel `ffo` — viés).
- Multi-tenancy B2B com regras de login e SSO por organização ([HN 39465740](https://hn.algolia.com/api/v1/items/39465740), `mffap` — funcionário, viés).
- Combinação com motores ReBAC elogiada por terceiros (`jzelinskie`, fundador da AuthZed, HN 31408059).

## Do que a comunidade se queixa (com evidência)
- **Volume e gravidade de falhas de segurança em 2026** (ver abaixo); relato "Why We Stopped Recommending ZITADEL for Self-Hosting" (dev.to, não lido; título via busca).
- **Preço do Cloud** e contagem de usuários ativos ([HN 42136000](https://hn.algolia.com/api/v1/items/42136000), `andix`: "very expensive"; dúvida sobre maturidade).
- **Complexidade de deploy** antiga (exigia CockroachDB/Kubernetes) (HN 31408059, `inportb`, `andy_ppp`); resolvida em parte com Postgres.
- **Bancos**: issue mais votada é "Support for more databases" #3598 (74 reações), agravada pela remoção do CockroachDB.
- Pedidos antigos sem atender ([issues por reações](https://github.com/zitadel/zitadel/issues?q=is%3Aissue+is%3Aopen+sort%3Areactions-%2B1-desc)): User Schema #6433 (62), autorização por grupos #5822 (35), ForwardAuth #6912 (23), remover campos obrigatórios #4386 (24), DPoP #5402, impersonação #7208.
- Consistência eventual visível ao usuário de API (issue #9599) e token invalidation incompatível com refresh silencioso (#9331).

## Histórico de segurança
~80 advisories no GitHub (8 páginas em 2026-11-03), a maioria sem CVE exibido; ≥ 20 High/Critical entre jun e out/2026.

| ID | Ano | Classe | Lição |
|---|---|---|---|
| GHSA-jh92-5mrj-p2w2 | 2026 | Login V2 confiava em ID de sessão de cookie não assinado e usava credencial própria `session.write` (confused deputy, CWE-441/565); IDs ordenados no tempo | UI de login não pode ser deputado privilegiado; estado no cliente sempre com integridade; IDs imprevisíveis não substituem autorização |
| GHSA-g8gj-gq47-xgf4, GHSA-738m-7888-jfv8, GHSA-jh3m-cr2x-qp88, GHSA-x4c7-fpcx-w9q6 | 2026 | Tomada de conta via IdP externo: vinculação manual, callback forjado, intent token forjável, confusão de IdP SAML | Vínculo de identidade federada é o ponto mais frágil; amarrar estado à sessão do navegador e ao IdP exato |
| GHSA-45f2-5q3r-xgg6, GHSA-pq2q-2c6r-75c4 | 2026 | Tomada de conta via cadastro de passkey; códigos de enrollment emitidos entre organizações | Cadastro de novo fator exige autenticação forte e checagem de tenant |
| GHSA-3gwm-5wx8-4gm6, GHSA-9993-rfwp-rhwf, GHSA-cfjq-28r2-4jv5 | 2025–2026 | Bypass de OTP e de MFA por reutilização de sessão | Máquina de estados de login explícita e testada por propriedade |
| GHSA-qj3q-235g-qr6w | 2026 | Tomada de conta em um clique via device code | Device flow precisa de confirmação contextual (phishing de código) |
| GHSA-mchc-246r-6m76, GHSA-mwmh-7px9-4c23 | 2025–2026 | Injeção de cabeçalho (Forwarded/Host) | Nunca derivar URLs/tenant de cabeçalhos sem lista de proxies confiáveis |
| GHSA-cpf4-pmr4-w6cx, GHSA-6x8v-2fq5-2229, GHSA-j344-gqv4-84ff | 2025–2026 | IDOR entre tenants, vazamento por IDs reciclados, enumeração entre organizações | Escopo de tenant em toda consulta; testes automatizados de isolamento |
| GHSA-7wfc-4796-gmg5, GHSA-93hm-8q29-c8cr | 2025–2026 | SSRF (Login V2; verificação de domínio) | Cliente HTTP de saída com bloqueio de rede interna por padrão |
| GHSA-fgmf-7rf8-m6vf | 2026 | Escape de sandbox do Actions V1 (JS embutido lê arquivos) | Evitar executar código de cliente no processo; preferir webhooks |
| GHSA-558c-v5wc-9w4q | 2026 | Usuários de organização desativada ainda logam (Login V2) | Reescrita de UI de login reintroduz regras antigas: testes de regressão de política |

Observação: a v3 chegou ao EOL em 31/ago/2026 e não recebeu correção do GHSA-jh92 (advisory). Muitas falhas surgiram com a coexistência Login V1 + Login V2.

## Arquitetura e decisões notáveis
- Monólito Go escalável horizontalmente, stateless (sem Redis para sessão), PostgreSQL como única dependência; cache opcional.
- Event store + projeções (v1–v4) migrando para relacional (v5).
- Login separado do núcleo (Login V2 em Next.js, consumindo a API pública): bom para customização, mas ampliou a superfície (confused deputy, SSRF, XSS).
- Actions v1 (JS embutido) → Actions v2 (chamadas HTTP a "targets" em "executions").

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- **Multi-tenancy**: instâncias por domínio dentro de um deployment; organizações B2B com políticas, branding, IdPs e domínios próprios; projetos concedidos entre organizações.
- **Extensibilidade**: Actions v2 (webhooks síncronos e assíncronos), Session API para UIs próprias.
- **Configuração**: arquivo/env para o runtime; recursos via API; políticas herdadas (instância → organização).
- **Bancos**: só PostgreSQL ≥ 14 (v3+); CockroachDB removido (PR #9444) por custo de manutenção e matriz de testes.
- **Soberania**: self-hosted ou Cloud com escolha de região; multi-região recomendada via clusters independentes + réplicas Postgres.

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **não**. Não há "approval"/quórum na doc (`apps/docs/content`) nem entre as issues mais votadas; autorização se limita a roles/grants de projeto.
- **Autorização sensível ao estado / transições**: **não**. RBAC por projeto; a issue #5822 (autorização por grupos) segue aberta. Não encontrado.

## Implicações para o projeto
1. **Evitar event sourcing como fonte de verdade** do AuthN; adotar estado relacional + evento de auditoria/outbox na mesma transação (lição direta de #9599 e do post de set/2026).
2. **Adotar** hierarquia instância/organização/projeto como inspiração para a seção 4.7, com testes automáticos de isolamento entre tenants.
3. **Segurança**: fluxos de vínculo de IdP, cadastro de passkey e device code merecem modelagem de ameaças e testes de propriedade; UI de login sem privilégios próprios.
4. **Evitar** suportar e depois remover bancos: escolher poucos bancos com matriz de testes completa (premissa 9 exige substituibilidade, mas com custo explícito).
5. **Licença**: AGPL afasta parte do mercado; Apache-2.0/MIT no SDK e no login é um padrão a considerar se adotarmos licença recíproca no núcleo.
6. **Oportunidade**: DPoP, PAR, FAPI, CIBA, códigos de recuperação e MFA adaptativo ainda ausentes.
7. **Adotar** política de suporte de versões mais longa que a da Zitadel (v3 sem patch crítico um mês após EOL).

## Fontes
- https://github.com/zitadel/zitadel (README, LICENSING.md) — 2026-11-03
- https://github.com/zitadel/zitadel/releases — 2026-11-03
- https://github.com/zitadel/zitadel/security/advisories (páginas 1, 2 e 5) — 2026-11-03
- https://github.com/zitadel/zitadel/security/advisories/GHSA-jh92-5mrj-p2w2 — 2026-11-03
- https://github.com/zitadel/zitadel/issues/9599 — 2026-11-03
- https://github.com/zitadel/zitadel/issues?q=is%3Aissue+is%3Aopen+sort%3Areactions-%2B1-desc — 2026-11-03
- https://github.com/zitadel/zitadel/pull/9444 e https://github.com/zitadel/zitadel/issues/9423 (via busca) — 2026-11-03
- Repositório, `apps/docs/content` (eventstore, architecture/solution, guides/integrate/dynamic-client-registration, back-channel-logout, self-hosting/manage/production) (clone esparso) — 2026-11-03
- https://zitadel.com/blog/apache-to-agpl (via busca) — 2026-11-03 — viés: fornecedor
- https://zitadel.com/license-faq (via busca) — 2026-11-03 — viés: fornecedor
- https://zitadel.com/blog/zitadel-v3-announcement (via busca) — 2026-11-03 — viés: fornecedor
- https://zitadel.com/blog/why-zitadel-nextgen-is-no-longer-event-sourced (via busca; não lido integralmente) — 2026-11-03 — viés: fornecedor
- https://dev.to/nirajkvinit1/why-we-stopped-recommending-zitadel-for-self-hosting-a-developers-field-report-4fdl (só título; domínio bloqueado) — 2026-11-03
- https://hn.algolia.com/api/v1/items/31408059 (HN "Zitadel: The best of Auth0 and Keycloak combined", 2022) — 2026-11-03 — vários comentários de funcionários e concorrentes
- https://hn.algolia.com/api/v1/items/42136000 e https://hn.algolia.com/api/v1/items/39465740 — 2026-11-03
