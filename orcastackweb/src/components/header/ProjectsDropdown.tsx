import type { HeaderNavigationItem } from '../../api';
import { HeaderDropdownContent } from './HeaderDropdownContent';

type Props = { items: HeaderNavigationItem[]; onNavigate: (path: string) => void };

export default function ProjectsDropdown({ items, onNavigate }: Props) {
  return <HeaderDropdownContent items={items} onNavigate={(_, path) => onNavigate(path)} section="projects" />;
}