import {
  Building2,
  FolderGit2,
  GitPullRequest,
  Rocket,
  Settings2,
  ShieldCheck,
  UserRound,
  UsersRound,
  Workflow,
} from 'lucide-react';
import { Link } from 'react-router-dom';

import type { HeaderNavigationItem, HeaderSection } from '../../api';

const itemIcons = {
  'building-2': Building2,
  'folder-git-2': FolderGit2,
  'git-pull-request': GitPullRequest,
  rocket: Rocket,
  'settings-2': Settings2,
  'shield-check': ShieldCheck,
  'user-round': UserRound,
  'users-round': UsersRound,
  workflow: Workflow,
};

type HeaderDropdownContentProps = {
  items: HeaderNavigationItem[];
  section: HeaderSection;
  onNavigate: (section: HeaderSection, path: string) => void;
};

export function HeaderDropdownContent({ items, section, onNavigate }: HeaderDropdownContentProps) {
  if (items.length === 0) {
    return <p className="enterprise-dropdown__empty">No actions are available for your current role.</p>;
  }

  return (
    <div className="enterprise-dropdown__items">
      {items.map((item) => {
        const Icon = itemIcons[item.icon as keyof typeof itemIcons] ?? Settings2;
        return (
          <Link className="enterprise-dropdown__item" key={item.id} onClick={() => onNavigate(section, item.path)} to={item.path}>
            <Icon aria-hidden="true" size={18} strokeWidth={1.8} />
            <span>
              <strong>{item.label}</strong>
              <small>{item.description}</small>
            </span>
            {item.count !== undefined ? <em>{item.count}</em> : null}
          </Link>
        );
      })}
    </div>
  );
}