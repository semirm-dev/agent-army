import { Link, useLocation } from 'react-router-dom';
import { Package, ClipboardList, RefreshCw, Stethoscope } from 'lucide-react';
import { cn } from '@/lib/utils';

const navItems = [
  { path: '/catalog', label: 'Catalog', icon: Package },
  { path: '/manifest', label: 'Manifest', icon: ClipboardList },
  { path: '/sync', label: 'Sync', icon: RefreshCw },
  { path: '/doctor', label: 'Doctor', icon: Stethoscope },
];

export function Sidebar() {
  const location = useLocation();

  return (
    <aside className="w-56 border-r bg-muted/30 p-4 flex flex-col gap-1">
      <div className="mb-6 px-2">
        <h1 className="text-lg font-bold">Agent Army</h1>
        <p className="text-xs text-muted-foreground">Plugin & Skill Manager</p>
      </div>
      <nav className="flex flex-col gap-1">
        {navItems.map((item) => {
          const Icon = item.icon;
          return (
            <Link
              key={item.path}
              to={item.path}
              className={cn(
                'flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors',
                location.pathname === item.path
                  ? 'bg-primary/10 text-primary font-medium'
                  : 'text-muted-foreground hover:bg-muted hover:text-foreground'
              )}
            >
              <Icon className="size-4" />
              {item.label}
            </Link>
          );
        })}
      </nav>
    </aside>
  );
}
