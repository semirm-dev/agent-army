import { Link, useLocation } from 'react-router-dom';
import { Package, ClipboardList, RefreshCw, Stethoscope, Sun, Moon } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useTheme } from '@/hooks/use-theme';

const navItems = [
  { path: '/catalog', label: 'Catalog', icon: Package },
  { path: '/manifest', label: 'Manifest', icon: ClipboardList },
  { path: '/sync', label: 'Sync', icon: RefreshCw },
  { path: '/doctor', label: 'Doctor', icon: Stethoscope },
];

export function Sidebar() {
  const location = useLocation();
  const { theme, toggleTheme } = useTheme();

  return (
    <aside className="w-56 border-r border-border bg-card flex flex-col">
      {/* Brand */}
      <div className="px-4 py-5 border-b border-border">
        <div className="font-mono text-sm font-bold text-primary">$ agent-army</div>
        <div className="text-[11px] text-muted-foreground mt-0.5">Plugin & Skill Manager</div>
      </div>

      {/* Nav */}
      <nav className="flex-1 py-3 px-2 flex flex-col gap-0.5">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = location.pathname === item.path;
          return (
            <Link
              key={item.path}
              to={item.path}
              className={cn(
                'flex items-center gap-2.5 px-3 py-2 text-[13px] rounded-md transition-colors relative',
                isActive
                  ? 'text-primary font-medium bg-primary/10'
                  : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'
              )}
            >
              {isActive && (
                <div className="absolute left-0 top-1/2 -translate-y-1/2 w-[2px] h-4 bg-primary rounded-r" />
              )}
              <Icon className="size-4" />
              {item.label}
            </Link>
          );
        })}
      </nav>

      {/* Footer */}
      <div className="px-4 py-3 border-t border-border flex items-center justify-between">
        <span className="font-mono text-[11px] text-muted-foreground/50">v0.3.0</span>
        <button
          onClick={toggleTheme}
          className="size-7 rounded-md flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-colors"
          title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
        >
          {theme === 'dark' ? <Sun className="size-3.5" /> : <Moon className="size-3.5" />}
        </button>
      </div>
    </aside>
  );
}
