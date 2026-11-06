# ADR 0024: padrões adotados pelo AuthZ

- Status: aprovado (2026-11-06)

## Contexto

O brief 4.29 propõe o AuthZEN Authorization API 1.0 (final desde 12/01/2026)
como contrato entre aplicações e PDP e pede avaliar OAuth RAR (RFC 9396),
UMA 2.0, XACML e escopos OAuth. O AuthZEN tem perfis em rascunho, entre eles
o de solicitação e aprovação de acesso (`docs/market-study/standards/authzen.md`).
Premissa 2: padrões abertos em vez de protocolos próprios.

## Decisão

| Padrão | Uso |
|---|---|
| AuthZEN 1.0 | API de decisão: avaliação individual, em lote e busca (para listagem) |
| Perfil AuthZEN de solicitação e aprovação (Access Request and Approval) | Acesso dinâmico e M de N |
| OAuth RAR (RFC 9396) | Nos tokens do AuthN |
| Escopos OAuth | Permissões amplas |
| UMA 2.0 | Fora; rever se a adoção crescer |
| XACML | Só referência |

## Alternativas

- **UMA 2.0:** baixa adoção; fica fora e será revisto.
- **XACML como linguagem ou protocolo:** padrão histórico; usado só como
  referência de conceitos.

## Consequências

- AuthZEN é o contrato C7 (design, seção 8); suíte de conformidade antes de
  cada release (design, seção 11).
- A busca do AuthZEN é a interface da listagem paginada
  ([ADR 0018](0018-authz-listing.md)).
- Solicitações e M de N seguem o perfil de solicitação e aprovação
  ([ADR 0020](0020-authz-dynamic-access.md),
  [ADR 0021](0021-authz-multi-party-approval.md)); o perfil está em Draft 1 e
  pode mudar.
- Tokens leves: papéis e escopos amplos no token, permissões finas no PDP
  (brief 4.29).
- O AuthZ funciona com IdP externo, recebendo o sujeito pelo AuthZEN
  (premissa 8).
