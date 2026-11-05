# OpenID AuthZEN: Authorization API 1.0 e perfis

Acesso a todas as fontes: 2026-11-03. Observação: `openid.net` não estava acessível por fetch nesta
sessão; as afirmações sobre o site da OpenID Foundation vêm dos resultados de busca e de cópias
do texto no repositório oficial do grupo (`github.com/openid/authzen`) e em `openid.github.io/authzen`.

## Resumo

- **Authorization API 1.0 é Final Specification da OpenID Foundation desde 12/01/2026**
  (votação de 23/12/2025 a 06/01/2026). Final = sem novas revisões; extensões só por novos documentos.
- O contrato é pequeno: modelo SARC (Subject, Action, Resource, Context) → `decision` booleano
  + `context` livre. Linguagem de políticas, arquitetura e armazenamento do PDP ficam **fora** do escopo.
- Há um ecossistema crescente de perfis em rascunho (Draft 1, fev.–out./2026): solicitação e
  aprovação de acesso, obrigações, emissão de tokens OAuth, Token Exchange/ID-JAG, claims de
  autorização em JWT, e o par COAZ/COAZ-MCP (mapeamento de MCP para AuthZEN).
- Adoção: dezenas de fornecedores passaram nos interops; suporte nativo ainda aparece como
  **experimental** em projetos open source relevantes (OpenFGA, Keycloak).

## API 1.0: endpoints (ligação HTTPS normativa)

| Operação | Caminho recomendado | Obrigatório no metadata? | Notas |
|---|---|---|---|
| Access Evaluation | `POST /access/v1/evaluation` | sim (`access_evaluation_endpoint`) | uma decisão |
| Access Evaluations (lote) | `POST /access/v1/evaluations` | não (`access_evaluations_endpoint`) | ver semânticas abaixo; **não** é `/evaluations/batch` |
| Subject Search | `POST /access/v1/search/subject` | não | "quem pode ver o documento 123?" |
| Resource Search | `POST /access/v1/search/resource` | não | "quais documentos Alice pode ver?" |
| Action Search | `POST /access/v1/search/action` | não | "o que Alice pode fazer no documento 123?" |
| Metadata (descoberta) | `GET /.well-known/authzen-configuration[/<sufixo>]` | — | RECOMENDADO; aceita sufixo por tenant (ex.: `/tenant1`) |

Pontos do texto (fonte: `api/authorization-api-1_0.md`):
- **Versão no caminho:** endpoints 1.0 "SHOULD include `v1`". Versões futuras podem só **acrescentar**.
- **Transporte:** agnóstico; HTTPS+JSON normativo; gRPC/CoAP podem vir como perfis.
- **Autenticação da API:** fora do escopo; OAuth 2.0 RECOMENDADO.
- **Decisão:** `true` = permitido; se o PEP não entender o `context`, PODE rejeitar. `false` = MUST NOT prosseguir.
- **`context` da decisão:** livre (motivos, obrigações/advices, dicas de UI, step-up). Interoperabilidade
  desse conteúdo depende de perfis (ex.: Obligations).
- **Lote (`evaluations`):** valores padrão no topo herdados por cada item; `options.evaluations_semantic`:
  `execute_all` (padrão), `deny_on_first_deny`, `permit_on_first_permit`.
- **Busca:** resultado "SHOULD" corresponder a `decision: true` numa avaliação posterior, sem garantia;
  busca RECOMENDADA transitiva (grupo → membro). Paginação por token opaco (`page.token`,
  `page.limit`, resposta `page.next_token`, `count` opcional); sem snapshot atômico; parâmetros
  devem ser idênticos entre páginas.
- **Metadata:** `policy_decision_point` (identificador https, previne mix-up de PDP),
  endpoints, `capabilities` (URNs em registro IANA), `signed_metadata` (JWT que prevalece sobre o JSON).
- **Identificação de requisição:** cabeçalho de ID de requisição ecoado na resposta (seção "Request Identification").
- **Registros IANA criados:** PDP Metadata, PDP Capabilities, sub-namespace URN `authzen`, well-known.

## Perfis e documentos em rascunho (todos "Draft 1")

| Documento | Data | Autor(es) | O que define | Relevância para nós |
|---|---|---|---|---|
| Access Request and Approval Profile (ARAP) | 29/10/2026 | K. McGuinness | "negação solicitável" (`context.access_request` com `expires_at`), endpoint de solicitação, *task handle* assíncrono, estados da tarefa, reavaliação após aprovação, artefatos de vinculação (JWT `binding_token`), "Approver Eligibility and Separation of Duties" | Alta: base para solicitações/JIT. **Invariantes:** negação solicitável continua negação; aprovação não concede acesso; o PDP decide de novo. Workflows, quórum, linguagens de aprovação e caixa de entrada do aprovador estão **fora do escopo**; há só `progress.current_step/total_steps` para aprovações multietapa |
| Companions do ARAP: Catalog, Bulk, Callback, Actor Delegation | out./2026 | K. McGuinness | catálogo de itens solicitáveis, pedidos múltiplos, notificação de conclusão autenticada, cadeia de atores (`client.actor`, `client.source`) com verificação | Actor Delegation cobre o caso "agente pede acesso em nome de usuário" |
| Obligations Profile | 03/07/2026 | A. Babeanu (Indykite) | objeto Obligation; semântica de cumprimento em PERMIT, DENY e busca; tipos normativos `step-up`, `notification`, `session_termination`, `custom`; descoberta e negociação de tipos suportados | Alta: padroniza o que hoje cada motor põe livremente em `context` |
| OAuth 2.0 Token Issuance Profile | 29/10/2026 | O. Gazitt | AS externaliza a decisão "emitir ou não o token" a um PDP; resposta pode moldar o token | Alta para o par AuthN+AuthZ: o AuthN pode consultar o AuthZ na emissão |
| OAuth 2.0 Token Exchange Binding | 29/10/2026 | O. Gazitt | mapeia RFC 8693 (delegação, impersonação), identity chaining, ID-JAG e Transaction Tokens em avaliações AuthZEN; distingue a autoridade de quem pede da autoridade delegada | Alta para agentes |
| Authorization Claims Profile | 29/10/2026 | O. Gazitt | preenche `groups`/`roles`/`entitlements` (RFC 9068) via Resource Search | Média: "papéis amplos no token" vindos do AuthZ |
| COAZ Framework | 02/2026 | Olivier, Tulshibagwale | mapeamento declarativo de qualquer protocolo para SARC; expressões em CEL | Média |
| COAZ-MCP Binding | 13/02/2026 | Tulshibagwale, Olivier | mapeamento padrão de cada método MCP (tools, resources, prompts, tasks...) para AuthZEN; servidor pode declarar mapeamento por tool no input schema (CEL) | Alta: autorização fina de chamadas MCP em gateway ou servidor |

Não encontrado: perfil gRPC publicado; perfil de "consistency token" (o "new enemy problem" não é tratado
na 1.0). Procurado no README e na lista de documentos do repositório `openid/authzen`.

## Implementações conhecidas

| Implementação | Tipo | Estado do suporte | Fonte (viés) |
|---|---|---|---|
| OpenFGA | motor ReBAC, Go | **experimental** (`--experimentals=authzen`); `evaluation`, `evaluations`, busca **sem paginação**; descoberta por store; docs recomendam API nativa para apps e AuthZEN para gateways/IdPs | openfga.dev (oficial) |
| Keycloak | IdP | **experimental** desde 26.7.0 (maio/2026); 26.8 passa a exigir token de cliente confidencial | keycloak.org (oficial); skycloak.io (terceiro) |
| Topaz (Aserto) | motor, Go | nativo desde 0.32.54 (fev./2025), REST e gRPC | aserto.com (fornecedor, editor da spec) |
| Cerbos | motor | anuncia suporte à 1.0 final | cerbos.dev (fornecedor) |
| Gravitee AM | IdP/gateway | integração AuthZEN; OpenFGA em "technical preview" | documentation.gravitee.io (fornecedor) |
| OPA | motor | via proxy/plugin (ex.: `kanywst/opa-authzen-plugin`, comunidade) | pkg.go.dev |
| Bibliotecas | Go: `github.com/SCKelemen/authzen` (alvo: 1.0 final); Ruby: `action_policy-authzen` | comunidade | pkg.go.dev, rubydoc |
| Perfil governamental | NLgov Profile for AuthZEN 1.0.0 (Logius, Países Baixos) | publicado | gitdocumentatie.logius.nl |

Não encontrado: anúncio de suporte AuthZEN em SpiceDB e Cedar (busca web, 2026-11-03).

## Resultados de interop (authzen-interop.net)

| Cenário | Versão da spec | Participantes listados |
|---|---|---|
| Todo (1.0, 1.1, 1.0-id) | Draft 02 | 3edges, Rock Solid Knowledge, Aserto, AuthZForce, Axiomatics, Cerbos, Hexa, Kogito, OPA, Permit, PingIdentity, PlainID, SGNL, Topaz, Amazon Verified Permissions, OpenFGA, EmpowerID, Indykite, WSO2 |
| API Gateway (gateway + app, dois PEPs) | Draft 02 | RSK, SGNL, Aserto, AVP, Axiomatics, Cerbos, Hexa, OpenFGA, Ping, PlainID, Topaz, WSO2 |
| Search (subject/resource/action) | Draft 03 (mai./2025) | Apache KIE, Axiomatics, Cerbos, EmpowerID, Indykite, Ping, PlainID, Topaz, WSO2 |
| IdP/PDP (IdP usa Resource Search para montar claims do token) | Draft 04 (interop dez./2025) | SGNL, Apache KIE, Axiomatics, Cerbos, EmpowerID, Indykite, Permit, Ping, PlainID, Topaz, WSO2 |

- Amostras verificadas: OpenFGA, OPA e AVP 43/43 PASS (Todo 1.1); Cerbos 198/198 (Search);
  Ping 6/6 (IdP). Vários participantes rodam **adaptadores/proxies** (ex.: OPA via proxy, AVP e Ping
  via parceiros), não suporte nativo.
- Interop ≠ certificação: os resultados são de rascunhos anteriores à 1.0 final. Um deck do
  Identiverse 2026 cita perfil de **certificação de conformidade** para PDPs AuthZEN, que verifica
  protocolo e não a correção das decisões (fonte secundária; não confirmado no site da OIDF).
- O grupo recebeu o prêmio "Outstanding Project Recognition" na EIC 2026 (fontes: openid.net, Cerbos).

## Implicações para o projeto

1. **Adotar a 1.0 como contrato público do AuthZ** (`/access/v1/*` + `/.well-known/authzen-configuration/<tenant>`):
   é final, pequena e já tem PEPs (gateways, IdPs) prontos para consumir.
2. Implementar **as cinco operações** desde o início, inclusive busca com paginação real — OpenFGA não
   pagina; é diferencial barato.
3. Usar `signed_metadata` e `policy_decision_point` para evitar mix-up de PDP em multi-tenant.
4. Tratar **`context` como extensão tipada**: seguir o Obligations Profile (step-up, notification,
   session_termination) e o ARAP (negação solicitável) em vez de inventar formatos.
5. **M de N:** o ARAP deixa quórum, separação de funções e workflow para a implementação. Nosso
   diferencial (quórum com assinaturas WebAuthn vinculadas à ação) pode ser exposto **atrás** do
   ARAP: a negação solicitável abre a tarefa; o quórum roda no nosso serviço; o PDP reavalia.
6. **Estados/transições:** a spec não tem conceito de estado; modelar como `action` = transição e
   `resource.properties.state` = estado atual, e a lista de transições permitidas via **Action Search**.
7. **Consistência:** a 1.0 não tem token de consistência; prever extensão via `capabilities`
   (URN própria) até haver perfil padrão.
8. Acompanhar Token Issuance + Token Exchange Binding: permitem o AuthN consultar o AuthZ na emissão
   de tokens e em delegação de agentes sem API proprietária.

## Fontes

- https://raw.githubusercontent.com/openid/authzen/main/api/authorization-api-1_0.md — 2026-11-03 — texto-fonte (ramo `main`; pode conter correções editoriais posteriores ao Final)
- https://raw.githubusercontent.com/openid/authzen/main/README.md — 2026-11-03
- https://openid.net/authorization-api-1-0-final-specification-approved/ — 2026-11-03 (via busca) — OIDF
- https://lists.openid.net/pipermail/openid-specs-authzen/2025-December/000500.html — 2026-11-03 (via busca) — aviso de votação
- https://openid.github.io/authzen/authzen-access-request-approval-profile-1_0.html — 2026-11-03
- https://openid.github.io/authzen/authzen-access-request-actor-profile-1_0.html — 2026-11-03
- https://openid.github.io/authzen/authzen-obligations-profile-1_0.html — 2026-11-03
- https://openid.github.io/authzen/authzen-oauth-token-issuance-1_0.html — 2026-11-03
- https://openid.github.io/authzen/authzen-oauth-token-exchange-1_0.html — 2026-11-03
- https://openid.github.io/authzen/authzen-oauth-authorization-claims-1_0.html — 2026-11-03
- https://openid.github.io/authzen/authzen-coaz-framework-1_0.html — 2026-11-03
- https://openid.github.io/authzen/authzen-coaz-mcp-binding-1_0.html — 2026-11-03
- https://authzen-interop.net/docs/intro e páginas de cenário/resultado — 2026-11-03 — site mantido com contribuição da Aserto (editora da spec)
- https://openfga.dev/docs/interacting/authzen — 2026-11-03 (via busca) — oficial
- https://www.keycloak.org/2026/05/authzen-as-experimental-feature — 2026-11-03 (via busca) — oficial
- https://skycloak.io/blog/keycloak-authzen-pdp-openid-authorization-api/ — 2026-11-03 — terceiro (vende hospedagem Keycloak)
- https://www.aserto.com/blog/authzen-gartner-iam-and-native-support-in-topaz — 2026-11-03 — fornecedor
- https://www.cerbos.dev/blog/openid-authzen-is-official-cerbos-is-ready — 2026-11-03 — fornecedor
- https://documentation.gravitee.io/am/guides/authorization-engines/authzen — 2026-11-03 — fornecedor
- https://pkg.go.dev/github.com/SCKelemen/authzen — 2026-11-03
- https://gitdocumentatie.logius.nl/publicatie/api/authzen/1.0.0/ — 2026-11-03 — governo NL
- https://www.cerbos.dev/blog/auth-zen-wins-outstanding-project-recognition-at-eic-2026 — 2026-11-03 — fornecedor
- https://docs.google.com/presentation/d/1Z-g-e6TMUGDYW3PI2yF6rcphjb8SlgrO/htmlpresent — 2026-11-03 (via busca) — deck Identiverse 2026, não verificado diretamente
