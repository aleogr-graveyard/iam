# Autenticação e autorização de agentes de IA e servidores MCP: estado da arte

Acesso a todas as fontes: 2026-11-03. Estados de rascunhos consultados na API do IETF Datatracker.
Área muito volátil: dezenas de rascunhos individuais novos por mês; só os com patrocínio relevante
estão listados.

## Resumo

- O **MCP** tem especificação de autorização estável baseada em OAuth 2.1 (versão atual do protocolo:
  **2026-07-28**). O servidor MCP é **resource server**; o AS pode ser externo. É exatamente o papel do
  nosso AuthN (AS) e do AuthZ (decisões finas por tool).
- Para empresas, o caminho padrão virou **ID-JAG / Cross-App Access** (IdP corporativo autoriza o
  agente a falar com outro app), adotado como extensão oficial do MCP.
- Delegação "agente em nome de usuário" usa **Token Exchange (RFC 8693) + claim `act`**; propagação
  interna, **Transaction Tokens**; identidade de workload, **WIMSE/SPIFFE**.
- Não há ainda padrão específico de "identidade de agente" aprovado: a IETF criou o WG **agentproto**
  (ativo desde out./2026) e há perfis concorrentes (AI-Auth, AAuth etc.).
- AuthZEN fecha a parte de decisão: **COAZ-MCP** (mapeia chamadas MCP para AuthZEN) e perfis de
  Token Exchange/ARAP Actor Delegation.

## 1. Especificação de autorização do MCP (2026-07-28)

| Item | Requisito |
|---|---|
| Escopo | só transporte HTTP; STDIO usa credenciais do ambiente |
| Papéis | servidor MCP = OAuth 2.1 resource server; cliente MCP = cliente OAuth; AS separado ou embutido |
| Descoberta | servidor **MUST** publicar Protected Resource Metadata (RFC 9728); cliente usa `WWW-Authenticate: Bearer resource_metadata=...`; AS **MUST** ter RFC 8414 ou OIDC Discovery; cliente MUST suportar ambos |
| Registro de cliente | prioridade: **Client ID Metadata Documents** (SHOULD; `client_id` é uma URL), pré-registro, e **Dynamic Client Registration (RFC 7591) agora _deprecated_** |
| Fluxo | OAuth 2.1 com PKCE; `resource` (RFC 8707) **MUST** em autorização e token, com URI canônica do servidor |
| Mix-up | AS SHOULD enviar `iss` (RFC 9207); cliente MUST validar contra o emissor registrado antes de trocar o código (SEP-2468) |
| Credenciais de cliente | vinculadas ao emissor: MUST NOT reutilizar com outro AS (SEP-2352) |
| Tokens | sempre no cabeçalho `Authorization`; servidor MUST validar audiência; **proibido _token passthrough_** (servidor que chama API a montante precisa de token próprio) |
| Escopos e step-up | 401/403 com `scope` no `WWW-Authenticate`; `insufficient_scope` em tempo de execução; cliente acumula escopos (união) no step-up; `scopes_supported` = mínimo |
| Refresh | AS MUST rotacionar refresh tokens de clientes públicos; tokens de acesso curtos (SHOULD) |
| Extensões oficiais (`ext-auth`) | **OAuth Client Credentials** (M2M, sem usuário); **Enterprise-Managed Authorization** (`io.modelcontextprotocol/enterprise-managed-authorization`): IdP corporativo decide acesso a servidores MCP via ID-JAG |

Evolução: 2025-03-26 (primeira versão OAuth) → 2025-06-18 (RS separado, RFC 9728/8707) → 2025-11-25
(CIMD) → 2026-07-28 (DCR deprecated, `iss` obrigatório na validação, protocolo sem sessão).

## 2. Padrões e rascunhos relevantes

| Documento | Status (2026-11) | O que resolve | Patrocínio |
|---|---|---|---|
| RFC 8693 Token Exchange | Proposed Standard | delegação (`act`) e impersonação | — |
| RFC 9728 Protected Resource Metadata | Proposed Standard (abr./2025) | descoberta do AS pelo cliente/agente | — |
| RFC 8707 Resource Indicators | Proposed Standard | audiência do token | — |
| RFC 9449 DPoP | Proposed Standard (set./2023) | token vinculado a chave (agentes guardam tokens) | — |
| draft-ietf-oauth-identity-assertion-authz-grant-06 (**ID-JAG**, "Cross-App Access") | WG document (31/10/2026) | app/agente troca asserção do IdP (Token Exchange) por um ID-JAG e o resgata no AS do recurso (RFC 7523) | Okta, Ping |
| draft-ietf-oauth-identity-chaining-17 | **RFC Editor queue** | encadear identidade entre domínios de confiança | WG OAuth |
| draft-ietf-oauth-transaction-tokens-11 | WG consensus, aguardando write-up | contexto de usuário+workload ao longo da cadeia interna | WG OAuth |
| draft-ietf-oauth-client-id-metadata-document-03 | WG document (02/11/2026) | `client_id` como URL, sem registro prévio (base do MCP) | WG OAuth |
| draft-ietf-oauth-attestation-based-client-auth-12 | WG consensus | autenticação de cliente por atestação (instância de app/agente) | WG OAuth |
| draft-ietf-wimse-arch-08 | WG document | identidade de workload multi-sistema (SPIFFE e afins) | WG WIMSE |
| draft-klrc-aiagent-auth-03 ("AI-Auth") | individual, informacional (jul./2026) | **perfil de boas práticas** com WIMSE + OAuth existentes; não cria protocolo | Defakto, AWS, Zscaler, Ping, OpenAI, Okta |
| draft-oauth-ai-agents-on-behalf-of-user-02 | **expirado** | `requested_actor` + `actor_token` no fluxo de código | individual |
| draft-hardt-oauth-aauth-protocol-11 (**AAuth**) + família (headers, R3, budgets, events) | individual (set.–out./2026) | protocolo novo agente→recurso com HTTP Message Signatures; 5 modos (identidade do agente, 2, 3 e 4 partes) | D. Hardt |
| WG **agentproto** (IETF) | **ativo** (out./2026) | protocolos de comunicação de agentes; muitos rascunhos iniciais | IETF |
| AuthZEN COAZ-MCP, Token Exchange Binding, ARAP Actor Delegation | Draft 1 (2026) | decisão fina por tool/argumento; autorizar trocas de token; agente solicitando acesso | OpenID AuthZEN WG |

## 3. Padrões de arquitetura que emergem

1. **Agente = cliente OAuth com identidade própria** (CIMD ou registro), nunca usa a senha/sessão do usuário.
2. **Delegação explícita:** token com `sub` = usuário e `act` = agente (cadeia de atores aninhada);
   consentimento por escopo/RAR; prazo curto; DPoP.
3. **Sem passthrough:** cada salto obtém token com audiência própria (Token Exchange ou ID-JAG).
4. **Controle corporativo:** IdP decide quais agentes acessam quais apps (ID-JAG / extensão MCP EMA).
5. **Autorização fina fora do token:** gateway ou servidor MCP consulta PDP por chamada de tool
   (COAZ-MCP), com obrigações (step-up, notificação) e **aprovação humana** para ações sensíveis
   (ARAP: negação solicitável → aprovação → reavaliação).
6. **Auditoria da cadeia:** registrar usuário, agente, cliente, tool, argumentos relevantes e decisão.

## 4. Implicações para o projeto

| # | Implicação | Produto | Quando |
|---|---|---|---|
| 1 | AuthN como AS compatível com MCP: RFC 9728 (lado RS nos SDKs), RFC 8414, CIMD, RFC 8707, RFC 9207 `iss`, PKCE, rotação de refresh. DCR só por compatibilidade | AuthN | junto com OAuth 2.1 |
| 2 | Token Exchange com `act` e políticas de quem pode agir por quem (consulta ao AuthZ via AuthZEN Token Exchange Binding) | AuthN+AuthZ | camada de agentes |
| 3 | Suportar ID-JAG como emissor (IdP) e como AS do recurso: abre o caso corporativo do MCP | AuthN | após Token Exchange |
| 4 | SDK/middleware "MCP server guard": valida token, responde desafios com `scope`, chama o PDP via COAZ-MCP | AuthZ | AuthZ v1+ |
| 5 | M de N e aprovações humanas para ações de agentes: usar ARAP + Actor Delegation como interface; quórum WebAuthn como diferencial | AuthZ | camada de acesso dinâmico |
| 6 | Não apostar em AAuth/AgentID/etc. até haver adoção ou WG; manter-se em OAuth + WIMSE (linha do AI-Auth) | — | revisar a cada 6 meses |
| 7 | Agentes como tipo de principal de primeira classe no modelo de dados (dono humano responsável, limites, expiração) | AuthN+AuthZ | modelo de dados inicial |

## Fontes

- https://modelcontextprotocol.io/specification/versioning — 2026-11-03 — oficial (versão atual 2026-07-28)
- https://modelcontextprotocol.io/specification/2026-07-28/basic/authorization — 2026-11-03 — oficial
- https://modelcontextprotocol.io/specification/2026-07-28/basic/authorization/security-considerations — 2026-11-03
- https://modelcontextprotocol.io/specification/2026-07-28/changelog — 2026-11-03
- https://modelcontextprotocol.io/extensions/auth/overview e /enterprise-managed-authorization — 2026-11-03
- https://datatracker.ietf.org/api/v1/doc/document/<nome>/ (estados dos rascunhos listados) — 2026-11-03
- https://www.ietf.org/archive/id/draft-ietf-oauth-identity-assertion-authz-grant-06.txt — 2026-11-03
- https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-03.txt — 2026-11-03
- https://www.ietf.org/archive/id/draft-klrc-aiagent-auth-03.txt — 2026-11-03
- https://www.ietf.org/archive/id/draft-oauth-ai-agents-on-behalf-of-user-02.txt — 2026-11-03
- https://www.ietf.org/archive/id/draft-hardt-oauth-aauth-protocol-11.txt — 2026-11-03
- https://datatracker.ietf.org/group/agentproto/ (API: estado "active", 2026-10-09) — 2026-11-03
- https://forkast.news/the-ietf-published-a-draft-for-agent-identity-the-market-already-shipped/ — 2026-11-03 — imprensa cripto, opinativa
- https://workos.com/blog/oauth-on-behalf-of-ai-agents — 2026-11-03 — fornecedor
- https://docs.litellm.ai/docs/proxy/id_jag/overview — 2026-11-03 — fornecedor (adoção de ID-JAG)
- https://www.mintmcp.com/docs/cross-app-access — 2026-11-03 — fornecedor
- https://ssojet.com/blog/okta-cross-app-access-xaa-id-jag — 2026-11-03 — fornecedor concorrente; cita "25+ early adopters" do Okta (não confirmado)
- `authzen.md` deste diretório (COAZ-MCP, Token Exchange Binding, ARAP)
