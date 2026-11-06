# ADR 0002: projetos A, B e C como clientes de teste fictícios

- Status: aprovado (2026-11-06)

## Contexto

O brief (seção 3) descreve os projetos clientes A, B e C, e a pergunta 1 da
seção 9 pergunta o que é o A e quando ele precisa estar rodando. A premissa 13
põe a plataforma antes dos produtos clientes. O iam é o primeiro de vários
projetos de infraestrutura, e os projetos finais que vão consumi-lo ainda não
existem.

## Decisão

1. A, B e C são **fictícios e mínimos**. Existem só para testar o iam durante o
   desenvolvimento.
2. São os clientes de teste de `examples/` e dos cenários de referência
   (brief 4.20). Os cenários da seção 3 do brief (clínicas, loja B2C, uso
   separado do AuthZ) viram dados de teste.
3. Sem prazo.
4. **Nenhum consumidor real** por enquanto.

## Alternativas

- **Tratar o A como projeto real, com prazo** (pergunta 1 do brief): não há
  projeto final que ocupe esse papel hoje.

## Consequências

- Nenhuma data externa dirige os marcos; o ritmo é o da verificação
  automatizada ([ADR 0004](0004-milestones-0-and-1.md)).
- Os requisitos vêm do brief e dos cenários, não de um usuário. Os cenários de
  referência, carregados pelos testes ponta a ponta, são o substituto do
  consumidor.
- Os clientes de teste usam só a interface pública (premissa 12):
  `examples/` é módulo separado, sem acesso a pacotes `internal`, verificado no
  CI (design, seção 4.3).
- O cliente A (web BFF em Go) é o critério de conclusão do marco 1
  ([ADR 0004](0004-milestones-0-and-1.md)).
- Nenhum dado pessoal real no laboratório (brief 4.17).
