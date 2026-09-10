import { Braces, BookOpenText, Box, Download } from 'lucide-react';
import { Link } from 'react-router-dom';

import { auditPublicHub } from '../api';

const resources = [
  { id: 'documentation', title: 'Documentation', detail: 'Architecture, operations, and security guidance.', meta: 'Read the platform model', icon: BookOpenText, path: '/docs' },
  { id: 'api-reference', title: 'API references', detail: 'Contracts for repositories, delivery, devices, and identity.', meta: 'Build against stable APIs', icon: Braces, path: '/developer' },
  { id: 'sdk-downloads', title: 'SDK downloads', detail: 'Node.js and Python integration packages from the source tree.', meta: 'Extend the control plane', icon: Download, path: '/developer' },
];

export function DeveloperResourceDeck() {
  return (
    <section className="resource-deck">
      <div className="resource-deck__copy">
        <span className="eyebrow">Developer provisions</span>
        <h2>Feed every build with governed resources.</h2>
        <p>Documentation, API contracts, and SDK entry points stay attached to the same platform that executes delivery work.</p>
        <Box aria-hidden="true" size={34} />
      </div>
      <div className="resource-deck__cards">
        {resources.map(({ id, title, detail, meta, icon: Icon, path }) => (
          <Link className="resource-flip" key={id} onClick={() => void auditPublicHub('resource', id)} to={path}>
            <span className="resource-flip__inner">
              <span className="resource-flip__face resource-flip__front"><Icon aria-hidden="true" size={26} /><strong>{title}</strong><small>{meta}</small></span>
              <span className="resource-flip__face resource-flip__back"><strong>{title}</strong><span>{detail}</span><small>Explore resource</small></span>
            </span>
          </Link>
        ))}
      </div>
    </section>
  );
}
