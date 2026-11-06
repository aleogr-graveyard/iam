# ADR 0020: acesso temporário, solicitações, delegação e obrigações

- Status: aprovado (2026-11-06)
- Escopo: fora do marco 1 ([ADR 0025](0025-authz-milestone-1-scope.md)).

## Contexto

O brief 4.29 e 5.2 pedem acessos temporários e just-in-time, solicitações e
aprovações, delegação e obrigações. No cenário da clínica, um médico
compartilha um prontuário com outro por tempo limitado. O perfil AuthZEN de
solicitação e aprovação (Draft 1) padroniza a "negação solicitável" e a
reavaliação depois da aprovação (`docs/market-study/standards/authzen.md`). O
estudo registra o Keycloak aceitando hora vinda do cliente.

## Decisão

| Tema | Decisão |
|---|---|
| Acesso temporário | Relações com janela de validade, medida pelo relógio do servidor |
| Solicitação | Máquina de estados que gera uma relação temporária, alinhada ao perfil AuthZEN de solicitação e aprovação |
| JIT | Caso particular de solicitação |
| Delegação | Só de subconjunto do que o delegante tem; com prazo; revogável; válida enquanto o delegante mantiver a permissão; profundidade 1 por padrão |
| Obrigações | Fazem parte da decisão; aplicação que não entende uma obrigação nega |
| Step-up | Obrigação de step-up liga ao `acr` do AuthN |

## Alternativas

Não houve proposta concorrente para o conjunto. Ficaram de fora, pelos
limites da própria decisão:

- **validade pelo relógio do cliente**;
- **delegação sem limite** (de qualquer permissão, sem prazo ou encadeada);
- **obrigação ignorável** pela aplicação, que seria fail-open.

## Consequências

- Testes de abuso obrigatórios (brief 4.19): delegar mais do que se tem, usar
  delegação depois de o delegante perder a permissão, usar acesso temporário
  depois do prazo.
- Solicitações usam `shared/statemachine`, com testes baseados em modelo
  (todas as transições válidas e inválidas).
- O step-up depende da sessão com fatores carimbados
  ([ADR 0027](0027-session-factors-and-nist-passwords.md)).
- M de N fica dentro do serviço de solicitações
  ([ADR 0021](0021-authz-multi-party-approval.md)).
- Agentes agindo em nome de alguém usam esta delegação
  ([ADR 0023](0023-authz-machines-and-agents.md)).
