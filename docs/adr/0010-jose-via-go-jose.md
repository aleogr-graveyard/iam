# ADR 0010: exceção de segurança da premissa 24 cobre JOSE (`go-jose` v4)

- Status: aprovado (2026-11-06)

## Contexto

A premissa 24 pede o mínimo de dependências, com uma exceção de segurança
(brief 4.30): primitivas criptográficas, hash de senhas e analisadores com
armadilhas conhecidas usam bibliotecas consolidadas, nunca implementação
própria (premissa 2). JOSE/JWT (JWS, JWE, JWK) não estava na lista. O brief
4.10 prevê JWE quando houver dados sensíveis no token. O estudo de mercado
registra 5 vulnerabilidades no `go-jose` em 2026 e nenhuma biblioteca Go de
identidade com auditoria independente publicada.

## Decisão

1. A exceção de segurança da premissa 24 **cobre JOSE/JWT**.
2. Biblioteca: **`github.com/go-jose/go-jose/v4`, versão ≥ 4.1.8**.
3. Usada **só dentro da camada criptográfica** (`shared/crypto`).
4. Sempre com **lista explícita de algoritmos** aceitos.

## Alternativas

- **Implementação própria de JOSE:** criptografia e parsing próprios, contra a
  premissa 2 e com as armadilhas conhecidas do formato.
- **`golang-jwt`:** cobre JWS/JWT, mas não JWE.

## Consequências

- `depguard` e `forbidigo` proíbem importar `go-jose` fora de `shared/crypto`
  (design, seção 4.3).
- Piso de versão fixado; `govulncheck` em cada PR; cada atualização revisada.
- Testes de abuso obrigatórios: `alg: none` e confusão de algoritmo
  ([ADR 0012](0012-absolute-prohibitions.md), design, seção 11).
- Limites de tamanho e profundidade antes do parsing de tokens.
