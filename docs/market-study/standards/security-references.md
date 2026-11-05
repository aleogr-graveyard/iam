# Referências de segurança, certificações e Cyber Resilience Act

Acesso a todas as fontes: 2026-11-03. Itens legais são leitura técnica, não parecer jurídico (ver seção 7 do brief).

## 1. NIST SP 800-63-4 (Digital Identity Guidelines)

**Status:** final, publicada em **julho de 2025** (31/07/2025; páginas de 26/08/2025 com ajustes
editoriais). Substitui a 800-63-3. Quatro volumes: base (gestão de risco, "DIRM"), 63A (proofing),
63B (autenticação), 63C (federação). Recurso de apoio: "Implementation Resources for SP 800-63 Revision 4".

### Níveis

| Nível | Resumo (Rev. 4) |
|---|---|
| IAL1 | proofing remoto ou presencial com atributos centrais ("core attributes"); protege contra fraude em escala básica |
| IAL2 | remoto ou presencial, evidência mais forte; contra ataques em escala/direcionados, falsificação básica e engenharia social |
| IAL3 | presencial supervisionado por agente treinado + coleta de pelo menos uma biometria |
| AAL1 | um fator; reautenticação ≤ 30 dias |
| AAL2 | dois fatores; **deve oferecer opção resistente a phishing**; reautenticação 24 h total / 1 h inatividade |
| AAL3 | autenticador criptográfico resistente a phishing com **chave não exportável**; 12 h total / 15 min inatividade |
| FAL1 | asserção portadora; proteção contra injeção recomendada |
| FAL2 | uma RP por asserção; proteção contra injeção **obrigatória** (transação começa na RP); confiança pré-estabelecida |
| FAL3 | asserção *holder-of-key* ou autenticador vinculado; chaves estabelecidas manualmente |

### Mudanças relevantes para o AuthN

- **Senhas:** mínimo **15** caracteres quando fator único; **8** quando parte de MFA; aceitar ≥ 64;
  Unicode conta por code point; **proibido** exigir composição ou troca periódica (troca só com
  evidência de comprometimento); obrigatório comparar com lista de bloqueio (vazadas, dicionário,
  contexto) sobre a senha inteira; proibido dica e perguntas de segurança (KBA); sem truncamento.
- **Passkeys sincronizáveis:** Apêndice B normativo admite autenticadores sincronizáveis (chaves
  cifradas na nuvem com ≥ 112 bits, operação sempre local). **Não valem para AAL3** (chave exportável);
  podem atender AAL2.
- **SMS/telefone (PSTN):** continua **restrito**: exige alternativa para todos os usuários, aviso
  ao usuário, considerar sinais de risco (troca de SIM, portabilidade). Remove a proibição de números VoIP.
  **E-mail não pode** ser canal de autenticação fora de banda.
- AAL3 deixa de exigir "hardware" e passa a exigir **chave não exportável**; FIPS 140 nível 1 no autenticador.
- Novos tópicos: carteiras controladas pelo usuário, autenticadores via NFC/Bluetooth e conexão híbrida
  (QR + BLE), **sessões vinculadas ao dispositivo** (device-bound session credentials), requisitos de
  cookies, monitoramento contínuo de sessão, recuperação de conta revisada, PAD obrigatório para face e
  proibição de biometria de voz, prevenção de injeção digital e mídia forjada no proofing.

**Implicação:** a matriz "método → nível de garantia" do brief pode usar AAL/IAL/FAL da Rev. 4 como
classificação padrão e gerar **alertas** (premissa 4) quando um tenant configurar algo abaixo (ex.:
senha < 15 sem MFA, SMS sem alternativa).

## 2. OWASP ASVS 5.0

- **Lançado em 30/05/2025** (tag `v5.0.0`). 17 capítulos, **345 requisitos** (contados no repositório):
  L1 = 70, L2 = 183, L3 = 92. Níveis cumulativos (L2 inclui L1; L3 inclui L2).
- Identificadores no formato `v5.0.0-<cap>.<seção>.<req>`; recomenda-se citar a versão.

| Cap. | Tema | Req. | Relevância |
|---|---|---|---|
| V1 | Encoding and Sanitization | 30 | geral |
| V2 | Validation and Business Logic | 13 | geral |
| V3 | Web Frontend Security | 31 | login-ui, console |
| V4 | API and Web Service | 16 | APIs |
| V5 | File Handling | 13 | baixa |
| **V6** | **Authentication** (doc., senhas, MFA, fora de banda, criptográfico, IdP) | 47 | AuthN — central |
| **V7** | **Session Management** | 19 | AuthN |
| **V8** | **Authorization** (documentação, design geral, nível de operação) | 13 | AuthZ e console |
| **V9** | **Self-contained Tokens** | 7 | JWT |
| **V10** | **OAuth and OIDC** (cliente, RS, AS, RP OIDC, OP, consentimento) | 36 | AuthN — central |
| V11 | Cryptography | 24 | camada criptográfica |
| V12 | Secure Communication | 12 | TLS |
| V13 | Configuration | 21 | segredos, config |
| V14 | Data Protection | 13 | LGPD/GDPR |
| V15 | Secure Coding and Architecture | 21 | dependências, arquitetura |
| V16 | Security Logging and Error Handling | 17 | auditoria |
| V17 | WebRTC | 12 | não se aplica |

**Implicação:** a meta do brief (L3 + matriz de rastreabilidade requisito→teste) é viável: ~345
itens, dos quais V6/V7/V8/V9/V10 (122) são o núcleo do produto. Começar com L2 completo no Marco 1.

## 3. FAPI 2.0

- **Security Profile e Attacker Model: Final** em fevereiro de 2025 (análise formal da Universidade
  de Stuttgart). **Message Signing: Final** anunciado em 26/09/2025.
- Exigências típicas: PAR obrigatório, PKCE, *sender-constrained tokens* (mTLS ou DPoP), autenticação de
  cliente forte (`private_key_jwt` ou mTLS), sem fluxo implícito, `iss` na resposta (RFC 9207).
- Testes de conformidade e certificação das versões finais disponíveis desde julho de 2025.
- **Implicação:** construir o AuthN já com PAR + DPoP + `private_key_jwt` torna FAPI 2.0 um perfil
  de configuração (premissa 3), relevante para Open Finance Brasil.

## 4. Certificações

| Programa | Processo | Custo | Observações |
|---|---|---|---|
| OpenID Certification (OIDC OP/RP, FAPI, CIBA, logout etc.) | **autocertificação**: roda a suíte oficial de conformidade (open source), envia logs e declaração | Connect: US$ 700 (membro) / US$ 3.500 (não membro) por implantação; FAPI: US$ 1.000 / US$ 5.000 (OP e RP pagos à parte). Valores de páginas sem data; confirmar | **Isenção para projetos open source** se nenhum mantenedor principal é pago para o projeto; análise caso a caso. Programa de testes independentes com 4 provedores anunciado para 2026 |
| FIDO2 Server (Functional) | testes de conformidade com ferramenta da FIDO + interop; submissão à FIDO | **US$ 6.000 (membro) / US$ 9.000 (não membro)**; derivada US$ 1.000/1.500; interop sob demanda US$ 3.000 | servidor não tem nível de segurança (níveis L1–L3+ são para autenticadores) |
| AuthZEN | interops abertos; perfil de certificação de PDP citado em 2026 | não encontrado | ver `authzen.md` |

Fontes de custo: páginas oficiais (FIDO lida diretamente; OpenID via resultados de busca, pois
`openid.net` não estava acessível nesta sessão).

## 5. Cyber Resilience Act (Regulamento (UE) 2024/2847)

| Pergunta | Resposta |
|---|---|
| Gestão de identidade/acesso está no Anexo III? | **Sim.** Anexo III, **Classe I**, item 1: "Identity management systems and privileged access management software and hardware, including authentication and access control readers, including biometric readers" |
| O que conta como "identity management system"? | Regulamento de Execução (UE) 2025/2392 (OJ 01/12/2025): produtos que "provide mechanisms for authentication **or authorisation**", inclusive gestão de ciclo de vida de credenciais e **sistemas de gestão de acesso**; inclui expressamente **SSO, federação, OTP, software de autenticação e MFA**. Ou seja, **AuthN e AuthZ** caem na Classe I |
| Avaliação de conformidade (Classe I) | Art. 32: autoavaliação (módulo A) **só** se aplicar integralmente normas harmonizadas/especificações comuns/certificação europeia; caso contrário, exame UE de tipo (B+C) ou garantia de qualidade total (H) por organismo notificado. **Exceção FOSS:** fabricante de software livre do Anexo III pode usar módulo A se a documentação técnica for pública |
| Normas harmonizadas | em elaboração no CEN-CENELEC (mandato M/606), com trabalho específico para "identity management systems" (webinar CEN-CENELEC de 25/02/2026). Não encontrado: norma publicada |
| Datas | notificação de vulnerabilidades ativamente exploradas e incidentes graves (Art. 14) desde **11/09/2026**; organismos notificados desde 11/06/2026; aplicação plena **11/12/2027** |
| Open source steward (Art. 3(14) e 24) | pessoa jurídica que, sem ser fabricante, apoia de forma sistemática e sustentada o desenvolvimento de FOSS destinado a atividades comerciais. Obrigações: política de cibersegurança documentada e verificável (tratamento e divulgação de vulnerabilidades), cooperar com autoridades de fiscalização, notificar vulnerabilidades exploradas (Art. 14(1)) na medida em que participa do desenvolvimento e incidentes que afetem sua infraestrutura de desenvolvimento. Sem marcação CE; regime de multas não se aplica a stewards |
| Somos steward ou fabricante? | Se monetizarmos o produto (open core, suporte pago, versão paga) colocando-o no mercado da UE, somos **fabricante** da Classe I. Steward só se uma entidade separada (ex.: fundação) apoiar o projeto sem monetizá-lo. Serviço puramente gerenciado (IDaaS) tende a cair em NIS2 e não no CRA, salvo componentes distribuídos (SDKs, agentes, apps) — **validar com jurídico** |

**Implicações:** planejar desde já (1) SBOM CycloneDX; (2) processo de vulnerabilidades com prazo de
24 h para alerta inicial à ENISA/CSIRT; (3) período de suporte declarado; (4) atualizações de segurança
separadas das funcionais; (5) documentação técnica **pública** (habilita módulo A pela exceção FOSS);
(6) acompanhar a norma harmonizada de IMS do CEN-CENELEC e usar ASVS como base até lá.

## Fontes

- https://pages.nist.gov/800-63-4/sp800-63.html, sp800-63a.html, sp800-63b.html, sp800-63c.html — 2026-11-03 — oficial
- https://csrc.nist.gov/pubs/sp/800/63/4/final — 2026-11-03 (via busca) — oficial
- https://pages.nist.gov/800-63-4-Implementation-Resources/ — 2026-11-03 (via busca)
- https://github.com/OWASP/ASVS/tree/v5.0.0/5.0/en (arquivos `0x10`–`0x26`) — 2026-11-03 — contagens feitas por nós
- https://www.appknox.com/cyber-security-jargons/owasp-asvs — 2026-11-03 — fornecedor (data de lançamento)
- https://openid.net/fapi-2-message-signing-final-specification-approved/ — 2026-11-03 (via busca)
- https://openid.net/fapi2-0-final-conformance-tests-available/ — 2026-11-03 (via busca)
- https://openid.net/certification/fees/ e https://openid.net/fee-schedule-to-certify-your-implementation/ — 2026-11-03 (via busca; valores divergentes entre páginas)
- https://openid.net/certification/open-source-project-certification-policy/ — 2026-11-03 (via busca)
- https://itbrief.co.uk/story/openid-foundation-unveils-new-independent-testing-plan — 2026-11-03 — imprensa
- https://fidoalliance.org/fido-certification-fees/ — 2026-11-03 — oficial (lido diretamente)
- https://www.springlex.eu/en/packages/cra/cra-regulation/annex-3/ — 2026-11-03 — espelho do texto oficial (EUR-Lex não respondeu)
- https://www.springlex.eu/en/packages/cra/cra-regulation/article-24/, article-32/, article-71/ — 2026-11-03
- https://www.springlex.eu/en/packages/cra/tdpc-regulation/annex-1/ — 2026-11-03 — Regulamento de Execução 2025/2392
- https://eur-lex.europa.eu/eli/reg/2024/2847/oj — referência canônica (não acessada)
- https://www.cencenelec.eu/media/CEN-CENELEC/Events/Webinars/2026/2026-02-25-webinar_cra-standards-unlocked-cybersecurity-requirements-for-identity-management-systems.pdf — 2026-11-03 (via busca)
- https://startwithidentity.com/blog/cyber-resilience-act-for-identity-teams/ — 2026-11-03 — blog de terceiros
