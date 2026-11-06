# ADR 0017: PDP central, cache e avaliação local no SDK

- Status: aprovado (2026-11-06)
- Escopo: cache de decisões no marco 1; avaliação local no marco 2
  ([ADR 0025](0025-authz-milestone-1-scope.md)).

## Contexto

O brief 4.15 pede degradação planejada: aplicações continuam funcionando com a
plataforma fora do ar. O brief 4.29 manda negar em caso de falha; a resiliência
vem da avaliação local e do cache, não de liberar o acesso. A meta de
avaliação local é < 1 ms ([ADR 0011](0011-latency-and-scale-targets.md)).

## Decisão

1. **PDP central** + **pacote versionado de políticas no SDK**.
2. Decisões **sem relações** (papéis no token, atributos, contexto):
   avaliadas **localmente** no SDK.
3. Decisões **relacionais**: no PDP, com **cache de TTL curto configurável por
   aplicação**, invalidado por versão de política e por token de consistência.
4. **PDP fora do ar:** cache dentro do TTL; depois, **nega**.
5. Avaliador local **só em Go**; outras linguagens via WebAssembly em camada
   posterior, sem port.

## Alternativas

- **PDP em sidecar com réplica dos dados:** um processo e uma cópia das
  relações por aplicação, com custo e operação a mais.
- **PDP sempre central:** cada decisão depende da rede e da disponibilidade da
  plataforma.

## Consequências

- Fail-closed ([ADR 0012](0012-absolute-prohibitions.md)): expirado o TTL sem
  PDP, nega. Teste de falha obrigatório.
- Revogações chegam ao SDK com atraso de até um TTL, exceto com token de
  consistência ([ADR 0016](0016-authz-relation-sync-and-consistency.md)).
- O avaliador local e o central decidem igual para a mesma entrada (teste de
  propriedade).
- Nenhum núcleo reimplementado em outra linguagem (brief 4.3).
