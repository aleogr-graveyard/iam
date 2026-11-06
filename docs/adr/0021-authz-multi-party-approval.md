# ADR 0021: aprovação M de N

- Status: aprovado (2026-11-06)
- Escopo: fora do marco 1 ([ADR 0025](0025-authz-milestone-1-scope.md)).

## Contexto

O brief 4.29 propõe aprovação multipartes (M de N) com assinatura WebAuthn
vinculada à ação exata. O estudo (`docs/market-study/authz/differentials-check.md`)
conclui que M de N sozinho não é diferencial (existe em PAM e nuvem), mas a
combinação de serviço genérico, assinatura por aprovador sobre a ação canônica,
quórum por papéis, evidência verificável pelo executor e perfil AuthZEN não foi
encontrada. Limite técnico: o WebAuthn não prova que o autenticador exibiu a
ação.

## Decisão

| Tema | Decisão |
|---|---|
| Onde | Dentro do serviço de solicitações ([ADR 0020](0020-authz-dynamic-access.md)) |
| Quórum | M de N, composição por papéis, limite de negações, janela e expiração |
| Aprovadores | Pessoas distintas; o solicitante não aprova; SoD e conflito de interesse por política |
| Ação | Representação canônica (RFC 8785) |
| Assinatura | Desafio WebAuthn = hash da ação + nonce, com verificação de usuário |
| Execução | Por reavaliação com a aprovação, ou por pacote de evidências verificável pelo executor |
| Interface | Renderiza a mesma representação canônica; limitação do WebAuthn documentada |
| Emergência | Caminho configurável por ação |
| Operador único | Atraso com janela de cancelamento |
| Shamir/FROST | Camada separada |
| Posicionamento | Contra PIM e IGA, não contra PAM |

## Alternativas

- **Quórum sem assinatura por ação** (como em PIM e PAM): a aprovação não fica
  vinculada à ação exata nem é verificável pelo executor.
- **Prometer que o autenticador mostra a ação** ("what you see is what you
  sign"): as extensões de transação do WebAuthn não existem nos navegadores.

## Consequências

- Testes de abuso obrigatórios (brief 4.19): autoaprovação, mesma pessoa
  aprovando duas vezes, aprovação reaproveitada para outra ação, troca da ação
  depois da aprovação.
- Canonicalização (RFC 8785) e verificação de assinaturas são código crítico:
  testes de propriedade e fuzzing.
- Aprovações multipartes são área sensível (brief 4.28) e passam por auditoria
  externa antes de produção (brief 4.25).
- Assinaturas guardadas na auditoria como evidência.
- A interface é parte da garantia: o que a pessoa vê depende dela.
