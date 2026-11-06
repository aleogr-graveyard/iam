# ADR 0007: issuer por tenant no caminho da URL

- Status: aprovado (2026-11-06)

## Contexto

O brief 4.17 deixou em aberto (pergunta 14) se o issuer OIDC de cada tenant
fica num caminho ou num subdomínio. O laboratório roda no Cloud Run sob
`lab.aleogr.dev`, com DNS manual na Cloudflare e custo mínimo (premissa 17,
[ADR 0001](0001-lab-hosting.md)). O RP ID do WebAuthn é o host do AuthN, nunca
`aleogr.dev` (brief 4.17).

## Decisão

1. **Issuer por caminho:** `https://<host>/t/<tenant>`.
2. O issuer vem da **configuração por tenant**; subdomínio ou domínio próprio
   continuam possíveis depois, sem mudar o modelo.
3. **Teste de abuso obrigatório:** sessão e passkey nunca valem fora do próprio
   tenant.

## Alternativas

- **Subdomínio por tenant** (`<tenant>.<host>`): exige DNS e certificado
  curinga, que no GCP pedem um load balancer global (custo fixo mensal), e
  traz risco de tomada de subdomínio quando um tenant sai.

## Consequências

- Um host, um certificado e um registro DNS no laboratório.
- Todos os tenants dividem o mesmo RP ID do WebAuthn e o mesmo host de cookies.
  O navegador não separa tenants: **o isolamento é da aplicação**.
- Testes de abuso obrigatórios desde o marco 1 (design, seções 6 e 11):
  - passkey registrada num tenant é recusada em qualquer outro;
  - sessão de um tenant não vale em outro;
  - token emitido para outro tenant é recusado.
- Discovery e endpoints de cada tenant ficam sob o próprio issuer.
- Nenhum host fixo no código: issuer, RP ID e domínios de cookie vêm de
  configuração (brief 4.17).
