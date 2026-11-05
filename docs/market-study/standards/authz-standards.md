# Padrões de autorização além do AuthZEN: o que adotar e quando

Acesso a todas as fontes: 2026-11-03. Estados de rascunhos IETF consultados na API do Datatracker.

## Visão rápida

| Padrão | Status (2026-11) | Camada | Recomendação | Quando |
|---|---|---|---|---|
| Escopos OAuth (RFC 6749 §3.3) + JWT access token (RFC 9068) | padrão consolidado | autorização de **identidade/API grossa** | **Adotar** | Marco 1 (AuthN) |
| Resource Indicators (RFC 8707) + Protected Resource Metadata (RFC 9728) | Proposed Standard (2020; abr./2025) | audiência do token / descoberta | **Adotar** | Marco 1; obrigatório para MCP |
| OAuth RAR (RFC 9396) | Proposed Standard, mai./2023 | consentimento/autorização **por transação** | **Adotar** no AuthN; ponte para M de N | após OAuth básico; junto com FAPI |
| Token Exchange (RFC 8693) | Proposed Standard, jan./2020 | delegação/impersonação | **Adotar** | junto com agentes/serviços |
| Transaction Tokens (draft-ietf-oauth-transaction-tokens-11) | WG Consensus, aguardando write-up | propagação de contexto **dentro** do domínio de confiança | **Acompanhar → adotar** | quando houver cadeia de microsserviços nos clientes |
| Shared Signals Framework 1.0, CAEP 1.0, RISC 1.0 (OpenID) | **Final** desde set./2025 | eventos de sessão/conta entre partes | **Adotar** (transmissor e receptor) | camada posterior do AuthN (revogação contínua) |
| UMA 2.0 (Kantara) | Recommendation, jan./2018 | compartilhamento controlado pelo **dono do recurso** | **Não adotar como núcleo**; estudar ideias | só se houver caso B2C de compartilhamento |
| XACML 3.0 / ACAL 1.0 + XACML 4.0 (OASIS) | 3.0 Standard (2013); 4.0 CSD01 (fev./2026) | linguagem de política + arquitetura PEP/PDP/PIP/PAP | **Referência histórica**; não implementar | nunca como contrato principal |
| AuthZEN 1.0 | Final, jan./2026 | contrato PEP↔PDP | **Adotar** (ver `authzen.md`) | primeira versão do AuthZ |

## Escopos OAuth

- Grão grosso: "o que o cliente pode pedir em nome do usuário", não "o que o usuário pode fazer".
- Boas práticas: escopos estáveis e poucos; permissões finas no PDP; `scope` no desafio
  `WWW-Authenticate` (RFC 6750) para step-up incremental (padrão também exigido pelo MCP).
- Risco recorrente: "explosão de escopos" (escopo por recurso). RAR e AuthZEN resolvem melhor.
- **Decisão sugerida:** escopos = capacidades da API; papéis amplos como claims (RFC 9068 `roles`/`groups`,
  ver AuthZEN Authorization Claims Profile); decisões finas sempre no PDP. Alinha com a [BOA PRÁTICA] "tokens leves" do brief.

## OAuth Rich Authorization Requests (RFC 9396)

- Parâmetro `authorization_details` (array JSON tipado por `type`) em autorização, PAR, token e introspecção;
  campos comuns: `locations`, `actions`, `datatypes`, `identifier`, `privileges`.
- Uso típico: Open Finance (pagamento de R$ X para Y), assinatura de documento específico, consentimento granular.
- Obrigatório/esperado em perfis FAPI 2.0 e Open Finance Brasil (o brief cita FAPI no escopo 5.1/5.5).
- **Relação com M de N:** RAR já fornece uma "representação canônica da ação" que pode ser assinada;
  um aprovador consente sobre `authorization_details`. Bom candidato para o formato da ação vinculada.
- **Quando:** AuthN, após OAuth 2.1 + PAR; o AuthZ consome `authorization_details` como `action`/`resource` do AuthZEN.

## Token Exchange (RFC 8693) e derivados

- `subject_token` + `actor_token` → token novo com claim `act` (delegação) ou sem (impersonação).
- Base de: identity chaining entre domínios (draft-ietf-oauth-identity-chaining, **na fila do RFC Editor**),
  ID-JAG (draft-ietf-oauth-identity-assertion-authz-grant-06, WG document) e Transaction Tokens.
- AuthZEN Token Exchange Binding (Draft 1, out./2026) define como o AS pergunta ao PDP se a troca é permitida.
- **Quando:** obrigatório para agentes e contas de serviço (domínios 5.1.6 e 5.2.8).

## Transaction Tokens (Txn-Tokens)

- Token curto, emitido por um serviço interno de Txn-Token, que carrega identidade do usuário, do
  workload e contexto de autorização **ao longo da cadeia de chamadas** dentro de um domínio de confiança;
  evita repassar o access token externo.
- Status: draft-ietf-oauth-transaction-tokens-11 (21/08/2026), estado "WG Consensus: Waiting for Write-Up".
- **Quando:** depois do AuthN básico; útil para clientes com microsserviços e para agentes que chamam
  várias ferramentas. Implementação é barata se Token Exchange já existir.

## Shared Signals Framework (SSF), CAEP e RISC

- SSF 1.0: transporte de Security Event Tokens (SET, RFC 8417) por push/poll, com gestão de streams;
  sujeitos por RFC 9493. CAEP 1.0: eventos de sessão (revogada, mudança de credencial, de conformidade
  do dispositivo, de nível de garantia). RISC 1.0: eventos de conta (comprometida, desabilitada etc.).
- **Final em setembro de 2025** (OpenID Foundation).
- Relevância: revogação contínua (logout global real), integração com IdPs externos e com o AuthZ
  (invalidar cache de decisões ao receber evento). Também é o canal natural para o AuthZ reagir a
  mudanças de risco.
- **Quando:** AuthN, camada de sessões avançadas; o AuthZ como receptor para invalidação de cache.

## UMA 2.0

- Grant OAuth com *permission ticket*: o dono do recurso define políticas no AS; uma "requesting party"
  obtém acesso assíncrono. Duas specs: UMA Grant e Federated Authorization (registro de recursos).
- Adoção baixa fora de nichos (saúde, governo, Keycloak Authorization Services implementa partes).
- **Recomendação:** não implementar como núcleo. Ideias aproveitáveis: compartilhamento iniciado pelo
  dono (ex.: paciente compartilha prontuário), *permission ticket* ≈ "negação solicitável" do AuthZEN ARAP.
  Reavaliar só se um projeto cliente B2C pedir compartilhamento entre usuários com consentimento.

## XACML

- Introduziu o vocabulário PEP/PDP/PIP/PAP, atributos, obrigações e advices, algoritmos de combinação
  (deny-overrides etc.). AuthZEN herda o vocabulário (cita XACML e NIST SP 800-162).
- XACML 4.0 (CSD01, 18/02/2026) reestrutura o padrão como ACAL 1.0 + representação XML; aprovado para
  revisão pública de 30 dias pelo TC em 17/02/2026. Não encontrado: estado posterior (procurado no
  site da OASIS e na busca web).
- **Recomendação:** referência conceitual (combinação de políticas, obrigações, separação PAP/PDP);
  não implementar XML/XACML. Clientes corporativos legados podem ser atendidos via AuthZEN (Axiomatics
  e PlainID já expõem AuthZEN).

## Ordem sugerida

1. Marco 1: escopos + RFC 9068 + RFC 8707/9728; AuthZ interno com contrato AuthZEN.
2. AuthZ v1: AuthZEN completo (evaluation, evaluations, 3 buscas, metadata).
3. AuthN avançado: PAR + RAR + Token Exchange; SSF/CAEP como transmissor.
4. Agentes: Token Exchange + ID-JAG + Txn-Tokens (ver `agents-mcp.md`).
5. Opcional por demanda: UMA (ideias), FAPI 2.0 completo (ver `security-references.md`).

## Fontes

- https://www.rfc-editor.org/rfc/rfc9396 (metadados via rfc9396.json) — 2026-11-03
- https://www.rfc-editor.org/rfc/rfc8693, rfc8707, rfc9068, rfc9728, rfc9700, rfc8417, rfc9493 — 2026-11-03
- https://datatracker.ietf.org/doc/draft-ietf-oauth-transaction-tokens/ (API: rev 11, 2026-08-21) — 2026-11-03
- https://www.ietf.org/archive/id/draft-ietf-oauth-transaction-tokens-11.txt — 2026-11-03
- https://datatracker.ietf.org/doc/draft-ietf-oauth-identity-chaining/ (rev 17, RFC Ed Queue) — 2026-11-03
- https://datatracker.ietf.org/doc/draft-ietf-oauth-identity-assertion-authz-grant/ (rev 06) — 2026-11-03
- https://openid.net/wg/sharedsignals/specifications/ — 2026-11-03 (via busca)
- https://www.helpnetsecurity.com/2025/09/22/openid-standards-real-time-security-event-sharing/ — 2026-11-03 — imprensa
- https://sgnl.ai/2025/09/sgnl-welcomes-the-publication-of-the-final-shared-signals-and-caep-specifications/ — 2026-11-03 — fornecedor (editor da spec)
- https://docs.kantarainitiative.org/uma/wg/rec-oauth-uma-grant-2.0.html — 2026-11-03 — Kantara
- https://docs.oasis-open.org/xacml/acal/xacml/core/v4.0/csd01/acal-core-xml-v4.0-csd01.html — 2026-11-03 (via busca)
- https://groups.oasis-open.org/discussion/minutes-for-17-february-2026-tc-meeting — 2026-11-03 (via busca)
- https://openid.github.io/authzen/authzen-oauth-token-exchange-1_0.html — 2026-11-03
