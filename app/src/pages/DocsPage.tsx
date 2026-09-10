type DocsPageProps = {
  onLogin: () => void;
  onSignup: () => void;
};

const docsCards = [
  {
    title: 'Platform architecture',
    description: 'Understand how repositories, gateway services, runner capacity, automation pools, and observability surfaces fit together.',
  },
  {
    title: 'API contracts',
    description: 'Work from a documented contract set for repository, pipeline, deployment, device, and access-management integration points.',
  },
  {
    title: 'Operations playbooks',
    description: 'Follow environment, rollout, and security procedures without leaving the product context that operators start from.',
  },
];

export function DocsPage({ onLogin, onSignup }: DocsPageProps) {
  return (
    <>
      <section className="page-hero">
        <span className="eyebrow">Docs</span>
        <h1>Documentation that matches the product surface.</h1>
        <p>
          The docs page frames ORCASTACK as a real operating platform, not just an API bundle. It orients new users around architecture, contracts, and deployment mechanics before they enter the dashboard.
        </p>
      </section>

      <section className="page-grid">
        {docsCards.map((card) => (
          <article className="page-card" key={card.title}>
            <h3>{card.title}</h3>
            <p>{card.description}</p>
          </article>
        ))}
      </section>

      <section className="cta-band">
        <div>
          <span className="eyebrow">Need access</span>
          <h2>Move from documentation to a managed workspace.</h2>
        </div>
        <div className="hero-actions">
          <button className="secondary-button" onClick={onLogin} type="button">Sign in</button>
          <button className="primary-button primary-button--warm" onClick={onSignup} type="button">Request access</button>
        </div>
      </section>
    </>
  );
}