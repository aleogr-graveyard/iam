# Verificação preliminar dos nomes candidatos (seção 10 do brief)

Data das verificações: 2026-11-03. **Preliminar:** não substitui busca de marca (INPI, EUIPO, USPTO) nem
consulta jurídica. Não foi feita busca em bases de marcas registradas.

## Método

| Verificação | Como | Leitura |
|---|---|---|
| `.com` | RDAP da Verisign (`rdap.verisign.com`) | 200 = registrado; 404 = provavelmente livre |
| `.dev` | RDAP do Google Registry (`pubapi.registry.google`) | idem |
| `.io` | RDAP da Identity Digital, só para parte dos nomes (limite de requisições) | idem |
| Pacotes | npm, PyPI, crates.io (nome exato) | ocupado/livre. Go não tem registro central: com caminho de importação próprio (vanity) não há conflito |
| GitHub | busca `user:<nome>` na API pública | usuário/organização com o nome exato |
| Conflitos de produto | busca web + conhecimento prévio (marcado "prévio" quando não verificado nesta sessão) | foco em segurança/identidade |

Resultado geral: **todos os 70 nomes de uma palavra têm o `.com` registrado e quase todos têm conta
homônima no GitHub.** Nome de uma palavra só será viável com sufixo/prefixo (ex.: `<nome>id`,
`<nome>-auth`, `get<nome>`) ou com domínio `.dev`/`.io` e organização no GitHub com variação.

## Tabela

Legenda: **V** = verde (sem conflito relevante encontrado, boa disponibilidade); **A** = amarelo
(conflito fraco ou disponibilidade baixa); **R** = vermelho (conflito no setor ou genérico demais).
Colunas de pacote: L = livre, O = ocupado. GitHub: U = usuário, Org = organização, livre.

### Pares em latim (AuthN / AuthZ)

| Nome | .com | .dev | .io | npm | PyPI | crates | GitHub | Conflitos / observações | Semáforo |
|---|---|---|---|---|---|---|---|---|---|
| Ipse | reg | reg | reg | L | L | L | U | nenhum produto de segurança encontrado; "IPSE" = associação de autônomos (UK) | V |
| Licet | reg | reg | reg | L | L | O | U | nenhum encontrado | V |
| Quis | reg | reg | — | O | O | L | U | genérico; difícil de buscar | A |
| Quid | reg | reg | — | O | O | O | Org | Quid (empresa de inteligência de mercado, prévio) | R |
| Sigillum | reg | livre? | — | L | O | O | U | longo | A |
| Placet | reg | reg | — | O | L | L | U | nenhum encontrado | A |
| Fides | reg | reg | — | O | O | O | Org | **Fides (Ethyca)**: plataforma open source de privacidade (Apache-2.0), setor adjacente | R |
| Facultas | reg | livre? | **livre** | L | L | L | **livre** | Facultas = editora universitária austríaca (prévio); nada em segurança | V |
| Tessera | reg | reg | — | O | O | O | Org | Google Tessera (logs de transparência); Tessera do ConsenSys (prévio) | R |
| Lex | reg | reg | — | O | O | O | U | Amazon Lex (prévio); genérico | R |
| Agnosco | reg | livre? | **livre** | L | L | L | U | nenhum encontrado; pronúncia/grafia menos óbvias | V |
| Permitto | reg | livre? | reg | L | L | L | U | nenhum encontrado | V |
| Verus | reg | reg | — | O | O | O | livre | **Verus** (verificador formal de Rust, CMU/Microsoft Research); blockchain Verus (prévio) | R |
| Fas | reg | reg | — | O | O | O | Org | sigla comum (FAS); impossível de buscar | R |
| Nomen | reg | reg | — | O | O | O | U | genérico | A |
| Mandatum | reg | reg | — | L | O | L | U | nenhum encontrado | A |
| Testis | reg | reg | — | O | L | L | U | conotação anatômica em inglês | R |
| Arbiter | reg | reg | — | O | O | O | U | muito usado em software | R |
| Identitas | reg | livre? | — | O | L | L | U | descritivo; fraco como marca | A |
| Auctoritas | reg | reg | reg | L | L | L | U | nenhum encontrado; longo | A |
| Signum | reg | reg | — | O | O | O | U | Signum (blockchain, prévio) | R |
| Ius | reg | livre? | — | O | L | L | U | curto, difícil de buscar | A |

### Plataforma em latim

| Nome | .com | .dev | .io | npm | PyPI | crates | GitHub | Conflitos / observações | Semáforo |
|---|---|---|---|---|---|---|---|---|---|
| Custos | reg | reg | reg | O | O | O | U | **Apache Airavata Custos: IAM open source** (Go, em reescrita) — conflito direto | R |
| Limen | reg | reg | reg | O | L | O | U | nenhum produto de segurança encontrado | A |
| Ianua | reg | reg | reg | L | O | L | U | nenhum encontrado | A/V |
| Arx | reg | reg | — | O | O | O | Org | ARX (anonimização de dados, Apache-2.0), setor adjacente | R |
| Fiducia | reg | reg | reg | L | O | L | U | Fiducia (TI bancária alemã, hoje Atruvia — prévio) | A |
| Tutela | reg | reg | — | L | L | L | U | Tutela (dados de rede móvel, prévio) | A |
| Clavis | reg | reg | — | O | O | O | U | **Clavis LSM** (proposta de módulo de segurança do kernel Linux para chaves) | R |
| Castellum | reg | livre? | — | L | O | O | U | Castellum (fundo imobiliário, prévio) | A |
| Vestibulum | reg | livre? | — | L | L | L | Org | palavra do *lorem ipsum*: buscabilidade péssima | R |
| Praesidium | reg | reg | — | L | O | L | U | Presidio (Microsoft, PII) | R |

### Pares em inglês (AuthN / AuthZ)

Todos com `.com` e `.dev` registrados; npm, PyPI e crates quase sempre ocupados (palavras comuns).

| Nome | Conflitos / observações | Semáforo |
|---|---|---|
| Avow / Allow | marca AVOW (verificação de identidade, Avow Networks); "allow" genérico | R |
| Attest / Assent | Attest (Org no GitHub; atestação é termo técnico: sigstore, WebAuthn) | R |
| Credence / Clearance | **Credence ID** (biometria/identidade digital) | R |
| Signet / Decree | Signet (bancos/joias, prévio); termos comuns | A |
| Sigil / Edict | SIGIL (protocolo para chamadas de agentes a ferramentas, crate Rust); Sigil (editor de e-books, prévio) | R |
| Emblem / Mandate | genéricos | A |
| Crest / Charter | **CREST** (organismo de acreditação em cibersegurança, prévio) | R |
| Witness / Warden | **Warden** (middleware de autenticação Ruby, base do Devise); witness (in-toto, prévio) | R |
| Herald / Steward | Herald (regras do Phabricator, prévio); "steward" colide com "open source steward" do CRA | R |
| Known / Granted | **Granted** (ferramenta de acesso a nuvem da Common Fate, prévio); genéricos | R |
| Seal / Writ | Microsoft SEAL (criptografia homomórfica, prévio) | R |
| Badge / Grant | genéricos ("grant" é termo do OAuth) | R |

### Plataforma em inglês

| Nome | .com | .dev | .io | Conflitos / observações | Semáforo |
|---|---|---|---|---|---|
| Gatehouse | reg | reg | — | genérico; vários produtos de controle de acesso físico (prévio) | A |
| Rampart | reg | reg | — | Apache Rampart (WS-Security) | R |
| Bulwark | reg | reg | reg | nenhum relevante encontrado; pacotes ocupados | A |
| Keep | reg | reg | — | Google Keep; Keep Network (fundiu-se na Threshold) | R |
| Barbican | reg | reg | — | **OpenStack Barbican** (segredos) | R |
| Citadel | reg | reg | — | **Istio Citadel** (CA de identidade de workloads); Citadel (finanças) | R |
| Stronghold | reg | reg | — | **IOTA Stronghold** (segredos) e Deckhouse Stronghold (compatível com Vault) | R |
| Garrison | reg | reg | reg | nenhum relevante encontrado | A |
| Archway | reg | reg | reg | Archway (blockchain Cosmos, prévio) | A |
| Covenant | reg | reg | — | Covenant (framework C2 ofensivo) | R |
| Threshold | reg | reg | — | **Threshold Network** (criptografia por limiar) — e colide com "assinatura por limiar" do próprio brief | R |
| Portico | reg | reg | reg | Portico (preservação digital, prévio) | A |
| Gatehall | reg | **livre?** | **livre** | npm/PyPI/crates livres; GitHub `GateHall` (usuário) | V |

## Repositório "iam"

| Aspecto | Achado |
|---|---|
| Pacotes | `iam` ocupado em npm, PyPI e crates; `.com`/`.dev` registrados; GitHub `iam` é usuário |
| Go | já há módulos `github.com/hanzoai/iam` (IdP/SSO Apache-2.0), `ranesidd/iam`, `theopenlane/iam` (cliente OpenFGA) |
| Buscabilidade | **muito ruim**: "IAM" é a categoria inteira e o nome dos serviços da AWS, GCP e Azure ("AWS IAM", "GCP IAM"); "iam" também é "I am" |
| Conflito de marca | não é marca (termo descritivo); sem risco jurídico, mas sem distinção |
| Recomendação | aceitável como **nome interno do diretório/repositório** se o caminho de importação for vanity (`<dominio>/iam/...`), pois o import path dá unicidade. Para o **nome público do repositório**, preferir o nome da plataforma (ex.: `gatehall/gatehall` ou `<org>/platform`) para ser encontrável |

## Top 5 recomendados

| # | Nome | Uso | Por quê | Próximo passo |
|---|---|---|---|---|
| 1 | **Agnosco / Permitto** | par AuthN / AuthZ | significados certos ("reconheço"/"permito"); `.dev` livre para os dois, `.io` livre para Agnosco; pacotes livres; nenhum conflito no setor | registrar `agnosco.dev`, `permitto.dev`, `agnosco.io`; buscar marcas |
| 2 | **Ipse / Licet** | par AuthN / AuthZ | curtos e distintos; pacotes livres (Licet ocupado só no crates); sem conflito no setor | domínios `.dev` e `.io` estão registrados: exigir variação (`ipse.id`?, `getipse.dev`) |
| 3 | **Gatehall** | plataforma | único da lista de plataforma com `.dev`/`.io` livres e pacotes livres; inglês claro, metáfora certa | registrar `gatehall.dev`/`.io`; organização GitHub com variação (`gatehall-dev`) |
| 4 | **Facultas** | AuthZ (alternativa a Permitto/Licet) | GitHub, `.io` e pacotes livres; "faculdade/permissão" | combinar com Agnosco ou Ipse |
| 5 | **Ianua** | plataforma (alternativa) | "porta"; sem conflito encontrado; npm e crates livres | domínios registrados: exige variação |

Combinação sugerida para discussão: **Gatehall** (plataforma) com **Agnosco** (AuthN) e **Permitto** (AuthZ).

Evitar, além da lista do brief: Custos, Fides, Clavis, Verus, Citadel, Stronghold, Threshold, Barbican,
Warden, Credence, Crest, Granted, Vestibulum.

## Fontes

- RDAP: https://rdap.verisign.com/com/v1/domain/<nome>.com, https://pubapi.registry.google/rdap/domain/<nome>.dev, https://rdap.identitydigital.services/rdap/domain/<nome>.io — 2026-11-03
- https://registry.npmjs.org/<nome>, https://pypi.org/pypi/<nome>/json, https://crates.io/api/v1/crates/<nome> — 2026-11-03
- https://api.github.com/search/users?q=user:<nome> — 2026-11-03
- https://airavata.apache.org/custos/ e https://github.com/apache/airavata-custos — 2026-11-03
- https://ethyca.com/open-source (Fides) — 2026-11-03 — fornecedor
- https://kernsec.org/pipermail/linux-security-module-archive/2025-March/049343.html (Clavis LSM) — 2026-11-03
- https://www.microsoft.com/en-us/research/publication/verus-verifying-rust-programs-using-linear-ghost-types/ — 2026-11-03
- https://bioinformaticshome.com/db/tool/arx/ (ARX) — 2026-11-03
- https://www.threshold.network/blog/what-is-threshold-cryptography — 2026-11-03
- https://blog.iota.org/iota-stronghold-beta-release/ e https://deckhouse.io/products/stronghold/ — 2026-11-03
- https://istio.io/v1.4/docs/concepts/security — 2026-11-03 (Citadel)
- https://github.com/wardencommunity/warden/wiki/overview — 2026-11-03
- https://trademark.justia.com/owners/avow-networks-incorporated-1640864 — 2026-11-03
- https://www.cbinsights.com/company/credence-id — 2026-11-03
- https://docs.rs/crate/sigil-protocol/latest — 2026-11-03
- https://pkg.go.dev/github.com/hanzoai/iam, https://pkg.go.dev/github.com/ranesidd/iam, https://beta.pkg.go.dev/github.com/theopenlane/iam — 2026-11-03
- Itens marcados "prévio": conhecimento anterior, não verificados nesta sessão (limite de buscas web atingido)
