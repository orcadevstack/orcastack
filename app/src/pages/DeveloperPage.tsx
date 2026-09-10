type DeveloperPageProps = {
  onLogin: () => void;
  onSignup: () => void;
};

const developerCards = [
  {
    title: 'SDK and API entry points',
    description: 'Developers can plug billing, orchestration, and pipeline workflows into consistent service contracts across the platform.',
  },
  {
    title: 'Runner-aware delivery',
    description: 'Build and release flows are designed around actual queue pressure, executor capacity, and hardware-backed validation lanes.',
  },
  {
    title: 'Policy-first automation',
    description: 'Release tooling stays attached to approvals, repository identity, and operational guardrails instead of bypassing them.',
  },
];

export function DeveloperPage({ onLogin, onSignup }: DeveloperPageProps) {
  return (
    <>
      <section className="page-hero">
        <span className="eyebrow">Developer</span>
        <h1>A developer page that feels tied to delivery reality.</h1>
        <p>
          This page gives the frontend a premium product narrative for engineers integrating with ORCASTACK. It highlights APIs, automation lanes, and operational constraints rather than generic marketing copy.
        </p>
      </section>

      <section className="page-grid">
        {developerCards.map((card) => (
          <article className="page-card" key={card.title}>
            <h3>{card.title}</h3>
            <p>{card.description}</p>
          </article>
        ))}
      </section>

      <section className="cta-band">
        <div>
          <span className="eyebrow">Developer access</span>
          <h2>Join a managed engineering workspace.</h2>
        </div>
        <div className="hero-actions">
          <button className="secondary-button" onClick={onLogin} type="button">Sign in</button>
          <button className="primary-button primary-button--warm" onClick={onSignup} type="button">Request access</button>
        </div>
      </section>
    </>
  );
}