import { Bell, LogOut, Settings2, ShieldCheck, UserRound } from 'lucide-react';
import { Link } from 'react-router-dom';

import type { HeaderNavigationItem } from '../../api';

const profileIcons = {
  bell: Bell,
  'log-out': LogOut,
  'settings-2': Settings2,
  'shield-check': ShieldCheck,
  'user-round': UserRound,
};

type ProfileDropdownProps = {
  items: HeaderNavigationItem[];
  onLogout: () => void;
  onNavigate: (path: string) => void;
};

export default function ProfileDropdown({ items, onLogout, onNavigate }: ProfileDropdownProps) {
  return (
    <div className="profile-dropdown__items">
      {items.map((item) => {
        const Icon = profileIcons[item.icon as keyof typeof profileIcons] ?? UserRound;
        const content = (
          <>
            <Icon aria-hidden="true" size={18} strokeWidth={1.8} />
            <span><strong>{item.label}</strong><small>{item.description}</small></span>
          </>
        );

        if (item.action === 'logout') {
          return <button className="profile-dropdown__item profile-dropdown__item--logout" key={item.id} onClick={onLogout} role="menuitem" type="button">{content}</button>;
        }

        return <Link className="profile-dropdown__item" key={item.id} onClick={() => onNavigate(item.path)} role="menuitem" to={item.path}>{content}</Link>;
      })}
    </div>
  );
}