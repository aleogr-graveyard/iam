# ADR 0006: API administrativa em Protobuf, servida com Connect, OpenAPI gerada

- Status: aprovado (2026-11-06)

## Contexto

O brief 4.3 propõe Protobuf/gRPC com `buf` para as APIs administrativas e pede
avaliar a convivência com OpenAPI/REST (pergunta 12). O brief 4.34 quer APIs
de protocolo implementadas exatamente como as especificações e uma API
administrativa consumível por SDKs gerados em várias linguagens.

## Decisão

1. **Protobuf é a fonte** da API administrativa (`api/`).
2. **Servida com Connect** (`connectrpc.com/connect`): gRPC e HTTP/JSON.
3. **OpenAPI gerada** a partir do Protobuf, para consumo REST.
4. **`buf breaking` no CI** (junto com `buf lint`): contrato C6.
5. **APIs de protocolo** (OIDC, OAuth, AuthZEN, WebAuthn) seguem as
   especificações e ficam fora do Protobuf.

## Alternativas

- **REST OpenAPI-first com a biblioteca padrão:** menos dependências, mas a
  fonte do contrato seria o YAML, sem a geração de clientes multilinguagem e a
  detecção de quebra que o `buf` oferece.
- **gRPC-Go + grpc-gateway:** HTTP/JSON por uma camada de tradução à parte,
  com mais peças e dependências; o Connect serve os dois a partir do mesmo
  handler.

## Consequências

- Dependências de runtime justificadas: `google.golang.org/protobuf` e
  `connectrpc.com/connect` (design, seção 10).
- Mudança incompatível na API administrativa falha no CI.
- Código gerado verificado como atualizado em cada PR (design, seção 11).
- SDKs gerados seguem a versão da API, não a do servidor
  ([ADR 0008](0008-single-product-version.md)).
- O marco 0 entrega um Protobuf mínimo passando por lint, breaking e geração
  ([ADR 0004](0004-milestones-0-and-1.md)).
