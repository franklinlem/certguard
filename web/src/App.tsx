export function App() {
  return <main>
    <header><a className="brand" href="/" aria-label="CertGuard, início"><span aria-hidden="true">◈</span> CertGuard</a><span className="phase">Em desenvolvimento</span></header>
    <section className="intro" aria-labelledby="title">
      <p className="eyebrow">MONITORAMENTO TLS • SELF-HOSTED</p>
      <h1 id="title">Confiança começa<br />pela verificação.</h1>
      <p className="description">Uma base segura para acompanhar a validade dos seus certificados. O motor TLS é o primeiro passo.</p>
    </section>
    <section className="foundation" aria-labelledby="foundation-title">
      <div><p className="eyebrow">ETAPA 01</p><h2 id="foundation-title">Fundação do CertGuard</h2><p>Inspeção de certificados com validação de hostname, cadeia de confiança e proteção contra acesso a redes privadas.</p></div>
      <ol aria-label="Etapas do produto">
        <li><span className="number">01</span><div><strong>Motor TLS</strong><span>Implementado · validação por testes automatizados</span></div></li>
        <li><span className="number">02</span><div><strong>Cadastro e histórico</strong><span>Próximo incremento · acesso autenticado</span></div></li>
        <li><span className="number">03</span><div><strong>Monitoramento e alertas</strong><span>Planejado · verificações periódicas</span></div></li>
      </ol>
    </section>
    <aside role="note"><strong>Ambiente de desenvolvimento.</strong> O monitoramento ainda não está ativo. Esta página não apresenta dados de certificados nem permite cadastrar domínios.</aside>
    <footer><span>CertGuard</span><span>Segurança antes de conveniência.</span></footer>
  </main>
}
