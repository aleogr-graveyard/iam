# Bibliotecas Go e Rust de identidade (zitadel/oidc, coreos/go-oidc, go-webauthn, go-jose, golang-jwt; openidconnect-rs, webauthn-rs)

| Campo | Valor |
|---|---|
| Categoria | Bibliotecas de protocolo/criptografia candidatas a **dependências** da plataforma (Go) ou referências de design (Rust) |
| Site / repositório | https://github.com/zitadel/oidc · https://github.com/coreos/go-oidc · https://github.com/go-webauthn/webauthn · https://github.com/go-jose/go-jose · https://github.com/golang-jwt/jwt · https://github.com/ramosbugs/openidconnect-rs · https://github.com/kanidm/webauthn-rs |
| Licença e modelo de negócio | zitadel/oidc Apache-2.0 (mantida pela ZITADEL, que promete manter "toda a funcionalidade aberta") · go-oidc Apache-2.0 · go-webauthn BSD-3-Clause · go-jose Apache-2.0 · golang-jwt MIT · openidconnect MIT · webauthn-rs **MPL-2.0** (copyleft por arquivo). Todas comunitárias, sem versão paga |
| Stack | Go (ver dependências abaixo); Rust (RustCrypto/OpenSSL) |
| Maturidade | Ver tabela "Panorama" |
| Acesso | 2026-11-03 |

> Método: versões e `go.mod` via proxy.golang.org; estrelas e OpenSSF Scorecard via deps.dev (Scorecard de 24/08/2026); vulnerabilidades via OSV (inclui Go Vulnerability Database) e páginas de advisories do GitHub; crates via API do crates.io. `api.github.com` não estava acessível na sessão.

## Resumo
- Para uma plataforma de segurança em Go, o núcleo **inevitável** é JOSE (não há JOSE na biblioteca padrão) e WebAuthn (formatos binários CBOR/atestação — exceção de segurança da 4.30). OIDC/OAuth é lógica de protocolo que o brief manda implementar.
- **go-jose v4** é a escolha de JOSE: API que obriga lista de algoritmos, Scorecard 6,8 com fuzzing, histórico só de DoS — mas 5 advisories em 2026.
- **go-webauthn** é a única opção Go madura para WebAuthn (fork mantido do duo-labs; Scorecard 8,5, fuzzing, sem advisories), porém v0.x com quebras sem aviso.
- **zitadel/oidc** é completo (OP + RP, RP certificado), mas puxa muitas dependências e teve uma falha criptográfica (tokens opacos sem autenticação) em 2026. **go-oidc** é mínimo (só RP) e bom para clientes de exemplo/SDKs.
- Nenhuma das bibliotecas Go tem **auditoria independente publicada** (não encontrada). A biblioteca padrão de criptografia do Go tem (Trail of Bits, 2025) e base FIPS 140-3.

## Panorama

| Biblioteca | Papel | Última versão | Estrelas | Scorecard (Maintained / Fuzzing / Code-Review) | Advisories (OSV/GHSA) | Auditoria |
|---|---|---|---|---|---|---|
| zitadel/oidc v3 | OP e RP OIDC/OAuth | v3.51.17 (02/11/2026) | 1.920 | 6,9 (10 / 0 / 10) | 1: GHSA-j8gq-92xf-382c (2026) | Não encontrada; RP **certificado OpenID** (perfis basic e config); OP não certificado |
| coreos/go-oidc v3 | RP: descoberta e verificação de ID token sobre `x/oauth2` | v3.21.0 (01/09/2026) | 2.502 | 5,2 (10 / 0 / 2) | 0 | Não encontrada |
| go-webauthn/webauthn | RP WebAuthn/passkeys | v0.18.2 (19/09/2026) | 1.356 | **8,5** (10 / 10 / 0) | 0 | Não encontrada; "FIDO2 Conformant" (testado com ferramentas de conformidade, autodeclarado) |
| go-jose/go-jose v4 | JWS/JWE/JWK/JWT | v4.1.8 (02/11/2026) | 543 | 6,8 (10 / 10 / 9) | 9 (v4), 11 (v3) | Não encontrada |
| golang-jwt/jwt v5 | JWT (só JWS) | v5.3.1 (28/01/2026) | 9.257 | 7,5 (**2** / 0 / 8) | 2 (v5); 4 (v4) | Não encontrada |
| openidconnect (Rust) | RP OIDC (tipado) | 4.1.0 (31/10/2026) | 648 | **3** (0 / 0 / 2) | 0 | Não encontrada |
| webauthn-rs (Rust) | RP WebAuthn | 0.6.2 (31/10/2026) | 716 | 5,9 (10 / 0 / 9) | 1: GHSA-22w3-693w-x895 (2026) | **Sim: SUSE Product Security** (README) |

## Dependências diretas (go.mod da última versão)
- **go-oidc**: `go-jose/v4` (≥ 4.1.4) e `golang.org/x/oauth2` — apenas 2. Atenção: o mínimo declarado (4.1.4) é anterior às correções de GHSA-mgcc-cww5-fp2q (4.1.6) e GHSA-4w3g/vvm8 (4.1.5); o consumidor precisa elevar go-jose.
- **zitadel/oidc**: 17 diretas (chi, gorilla/securecookie, rs/cors, otel, google/uuid, `zitadel/schema`, `muhlemmer/gu`, `muhlemmer/httpforwarded`, `bmatcuk/doublestar`, `google/go-github/v31`, `golang/mock`, testify…) + 9 indiretas; Go ≥ 1.26 (suporta só as duas últimas versões).
- **go-webauthn**: `fxamacker/cbor/v2`, `golang-jwt/jwt/v5` (blob do FIDO MDS3), `google/go-tpm` (atestação TPM), `go-webauthn/x`, `go-viper/mapstructure`, `tinylib/msgp`, `google/uuid`, `go.uber.org/mock`; toolchain go1.27.1.
- **go-jose v4**: nenhuma dependência externa (usa fork interno de `encoding/json` com casamento de nomes sensível a maiúsculas); v5 virá sobre `encoding/json/v2`.
- **golang-jwt v5**: nenhuma; Go ≥ 1.21.

## Cobertura por domínio (5.1) — blocos que cada biblioteca entrega

| # | Domínio | Suporte | Notas |
|---|---|---|---|
| 1 | Autenticação primária | ◐ | Passkeys (go-webauthn); login social/SSO OIDC como RP (go-oidc, zitadel/oidc). SAML, LDAP, ICP: fora do escopo destas libs |
| 2 | MFA e step-up | ◐ | WebAuthn (chaves físicas e plataforma) via go-webauthn. TOTP: nenhuma (implementação trivial com stdlib) |
| 3 | Ciclo de vida da conta | ○ | Fora do escopo |
| 4 | Sessões e tokens | ◐ | Assinatura/verificação de tokens (go-jose, golang-jwt); refresh/revogação no OP do zitadel/oidc. DPoP: não listado no zitadel/oidc |
| 5 | Protocolos | ◐ | zitadel/oidc OP/RP: code, implicit (OP), client credentials, refresh, discovery, JWT profile (RFC 7523), PKCE, token exchange (RFC 8693), device (RFC 8628), back-channel logout (OP); mTLS "not yet"; PAR, CIBA, FAPI não listados. openidconnect-rs: code, implicit, hybrid, device, introspecção, revogação |
| 6 | Máquinas e agentes | ◐ | Client credentials e JWT profile (zitadel/oidc) |
| 7–11 | Demais | ○ | Fora do escopo |

## Diferenciais / exclusivos
- **go-jose v4**: `ParseSigned*` exige lista não vazia de algoritmos aceitos — algoritmo-confusão bloqueado pela API, não pela disciplina do usuário.
- **golang-jwt**: `alg=none` só com a constante `UnsafeAllowNoneSignatureType`; README pede validar `alg` explicitamente.
- **zitadel/oidc**: interface `Storage` para o OP (o produto ZITADEL é construído sobre ela) — exemplo de portas e adaptadores.
- **go-webauthn**: tipos para extensões (PRF, largeBlob, credProps, credProtect), atestação packed/tpm/android-key/apple/compound, provedor de metadados (MDS3).
- **openidconnect-rs 4.x**: *typestates* garantem em tempo de compilação que o cliente só usa fluxos cujos endpoints foram configurados — ideia aplicável em Go com tipos distintos por estado.
- **webauthn-rs**: padrões opinativos (UV exigida, resident keys desencorajadas para não "entupir" chaves físicas, só algoritmos seguros) e migração de OpenSSL para RustCrypto (0.6).

## O que a comunidade elogia (com evidência)
- go-jose e go-oidc são dependências de fato do ecossistema cloud-native (Kubernetes/Dex/etc. — conhecimento prévio); Scorecard "Maintained 10" para ambos.
- go-webauthn: melhor pontuação de Scorecard do grupo (fuzzing e SAST 10).
- webauthn-rs: auditado pela SUSE e usado no Kanidm (README).
- Demanda por solução Go "pronta": "I really wish there was such an easy off-the shelf auth solution for Go" ([HN 45389293](https://news.ycombinator.com/item?id=45389293)).

## Do que a comunidade se queixa (com evidência)
- golang-jwt: Scorecard "Maintained 2/10" (sem release desde 01/2026; 55 issues abertas); nasceu de fork porque o `dgrijalva/jwt-go` foi abandonado com CVE aberto (CVE-2020-26160).
- go-webauthn: "version 0... there may be breaking changes without warning" (README).
- openidconnect-rs: Scorecard "Maintained 0" e "Security-Policy 0" em 08/2026 (antes da 4.1.0); 82 issues abertas; mantenedor único (conhecimento prévio).
- go-oidc: Code-Review 2/10 (commits sem revisão por segundo mantenedor).
- zitadel/oidc: grande número de dependências e exigência de Go muito recente.

## Histórico de segurança

| ID | Ano | Biblioteca / classe | Lição para nós |
|---|---|---|---|
| GHSA-j8gq-92xf-382c | 2026 | zitadel/oidc OP < 3.47.0 — **tokens de acesso opacos cifrados sem autenticação**: o portador altera o ciphertext e forja o `sub`, bypass completo via introspecção (CVSS 6,5) | Tokens opacos = identificador aleatório + lookup, ou AEAD/assinatura; nunca cifra maleável. Revisão criptográfica obrigatória mesmo usando primitivas da stdlib |
| GHSA-2c7c-3mj9-8fqh | 2023 | go-jose — PBES2 com contagem de iterações ilimitada (DoS) | Limitar parâmetros vindos do atacante (p2c, tamanho) |
| CVE-2024-28180 | 2024 | go-jose — bomba de descompressão em JWE | Limite de tamanho pós-descompressão |
| CVE-2025-27144 | 2025 | go-jose — parsing com consumo excessivo (DoS) | Limitar entrada antes de `split` |
| CVE-2026-34986 | 2026 | go-jose — *panic* em JWE com `encrypted_key` vazio | Fuzzing contínuo dos parsers (premissa 16) |
| GHSA-mgcc-cww5-fp2q | 2026 | go-jose < 4.1.6 — JWK RSA com módulo/primos sem limites → `Validate()` cúbico (DoS) | Limitar tamanho de chaves aceitas em JWKS de terceiros (federação) |
| GHSA-4w3g-jfj2-cw2q, GHSA-vvm8-gf78-28g4 | 2026 | go-jose < 4.1.5 — *panic* em CBC-HMAC; JSON profundamente aninhado estoura pilha | Rejeitar JOSE não usado (JWE/CBC) por configuração; limitar profundidade |
| CVE-2025-30204 | 2025 | golang-jwt — `strings.Split` em header malicioso aloca memória excessiva | Idem |
| CVE-2024-51744 | 2024 | golang-jwt v4 — semântica de erro de `ParseWithClaims` levava a checar só "expirado" e ignorar "inválido" | API deve tornar o uso errado difícil (um único "válido/ inválido") |
| CVE-2020-26160 | 2020 | dgrijalva/jwt-go — `aud` como lista permitia bypass; projeto abandonado | Critério de dependência: manutenção ativa e plano de saída |
| GHSA-22w3-693w-x895 | 2026 | webauthn-rs < 0.5.5 — origem validada com `ends_with` sem checar o ponto (subdomínio) | Comparação de origem/RP ID por rótulos DNS, nunca por sufixo de string |
| Adjacentes: CVE-2020-15216, CVE-2026-33487 (goxmldsig); CVE-2022-41912, CVE-2020-27846 (crewjam/saml) | 2020–26 | Bypass de assinatura XML (incluindo captura de variável de laço em 2026) | SAML em Go é o maior risco de dependência; exigir testes XSW próprios e acompanhamento ativo |
| Adjacente: lestrrat-go/jwx v2 | 2023–24 | Oráculo de padding, DoS em JWE, *panic* | Alternativa a go-jose com histórico semelhante |
| Contexto (produto, não lib): GO-2026-5768 / CVE-2026-55672 | 2026 | ZITADEL (sobre zitadel/oidc) — `client_id` não vinculado na troca de código/refresh | A lib entrega o protocolo, mas vinculações (código↔cliente) ficam no storage: testar no produto |

Go stdlib: auditoria da Trail of Bits (2025) nos pacotes `crypto/*` do módulo FIPS 140-3 encontrou 1 vulnerabilidade de baixa severidade (integração legada BoringCrypto) — [Go Blog](https://go.dev/blog/tob-crypto-audit).

## Arquitetura e decisões notáveis
- go-oidc delega OAuth a `golang.org/x/oauth2` (mantido pelo time do Go) e JOSE a go-jose: mínimo e composável.
- zitadel/oidc empacota servidor HTTP (chi), CORS, cookies e OTel: conveniente, mas impõe escolhas de infraestrutura e amplia o SBOM.
- go-jose mantém fork do `encoding/json` para evitar diferenças de interpretação entre linguagens (casamento sensível a maiúsculas) — decisão de segurança relevante para nós.
- Política de suporte: go-jose e zitadel/oidc suportam só as 2 últimas versões do Go; go-webauthn usa `toolchain` no go.mod.

## Adequação como dependências (premissas 2 e 24; política da 4.30)

| Biblioteca | Recomendação | Justificativa |
|---|---|---|
| Biblioteca padrão Go (`crypto/*`) + `golang.org/x/crypto` (argon2) | **Usar** | Auditada, FIPS 140-3, mantida pelo time do Go |
| go-jose v4 | **Usar** (com limites de entrada e versão ≥ 4.1.8) | Zero dependências, API segura por padrão, fuzzing; acompanhar advisories (cadência alta em 2026) e migração para v5 |
| go-webauthn/webauthn | **Usar** (fixar versão, testes de conformidade FIDO próprios) | Única opção madura em Go; parsers CBOR/atestação são exatamente a exceção de segurança da 4.30 |
| golang-jwt v5 | **Evitar** como dependência direta | Redundante com go-jose; manutenção baixa. (Entra transitivamente via go-webauthn para MDS3) |
| coreos/go-oidc v3 | **Usar em clientes de exemplo/SDKs Go**, não no servidor | Mínimo e estável; só RP |
| zitadel/oidc v3 | **Referência, não dependência** do OP | Brief manda implementar lógica de protocolo; árvore de dependências grande; falha criptográfica recente. Útil como referência de interface `Storage` e como RP certificado para testes de interoperabilidade |
| openidconnect-rs, webauthn-rs | **Referências de design** | Rust não entra no binário Go; usar como padrão de API (typestates, padrões opinativos) e webauthn-rs como contraparte de testes cruzados |

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Não se aplica às bibliotecas (sem persistência própria). zitadel/oidc permite múltiplos issuers por requisição (o produto ZITADEL é multi-instância — conhecimento prévio); go-webauthn aceita múltiplas origens/RP IDs por configuração.

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N)**: **não** (fora do escopo de todas).
- **Autorização sensível ao estado / transições declarativas**: **não** (fora do escopo).

## Implicações para o projeto
1. **Adotar** go-jose v4 e go-webauthn como as duas dependências de segurança do AuthN, com justificativa registrada (4.30), versão mínima por advisory e `govulncheck` no CI.
2. **Adotar** limites explícitos em toda entrada criptográfica (tamanho de token/JWK/JWE, profundidade JSON, parâmetros PBES2, tamanho de chaves RSA) — 8 dos 12 advisories das bibliotecas acima são DoS por entrada não limitada.
3. **Adotar** fuzzing nativo do Go (`go test -fuzz`) sobre nossos parsers e sobre as fronteiras com go-jose/go-webauthn.
4. **Evitar** tokens opacos "autocontidos" cifrados; preferir identificador aleatório com hash no banco (revogação trivial) ou JWT assinado.
5. **Implementar** OP/RP OIDC próprio (como manda a 4.30) usando zitadel/oidc e Spring AS como referência e certificar via OpenID Foundation conformance suite desde o primeiro marco.
6. **Oportunidade**: não existe solução Go "completa e fácil" de AuthN como biblioteca; nosso modo biblioteca preenche essa lacuna.
7. **Risco**: SAML em Go (goxmldsig/crewjam) tem histórico de bypass de assinatura — decidir cedo entre dependência auditada, implementação mínima com testes XSW extensivos, ou SAML via componente isolado.

## Fontes
- https://proxy.golang.org/{github.com/zitadel/oidc/v3, github.com/coreos/go-oidc/v3, github.com/go-webauthn/webauthn, github.com/go-jose/go-jose/v4, github.com/golang-jwt/jwt/v5, github.com/go-webauthn/x}/@latest e .mod — 2026-11-03
- https://api.deps.dev/v3/projects/ (os 7 repositórios) — 2026-11-03
- https://api.osv.dev (Go: módulos citados; crates.io: openidconnect, oauth2, webauthn-rs, webauthn-rs-core, webauthn-authenticator-rs; Go adjacentes: lestrrat-go/jwx, crewjam/saml, russellhaering/goxmldsig) — 2026-11-03
- https://github.com/zitadel/oidc (README) e https://github.com/zitadel/oidc/security/advisories/GHSA-j8gq-92xf-382c — 2026-11-03
- https://github.com/coreos/go-oidc e /security/advisories — 2026-11-03
- https://github.com/go-webauthn/webauthn e /security/advisories — 2026-11-03
- https://raw.githubusercontent.com/go-jose/go-jose/main/README.md e `jws.go` — 2026-11-03
- https://github.com/golang-jwt/jwt/security/advisories e README — 2026-11-03
- https://github.com/kanidm/webauthn-rs (README: auditoria SUSE) — 2026-11-03
- https://raw.githubusercontent.com/ramosbugs/openidconnect-rs/main/README.md e UPGRADE.md — 2026-11-03
- https://crates.io/api/v1/crates/{openidconnect, webauthn-rs, webauthn-rs-core} — 2026-11-03
- https://osv.dev/vulnerability/GO-2026-5768 — 2026-11-03
- https://go.dev/blog/tob-crypto-audit (via busca) — 2026-11-03 — mantenedores do Go
- https://news.ycombinator.com/item?id=45389293 — 2026-11-03
