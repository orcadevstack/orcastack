import type React from 'react';

type SectionProps = {
  title: string;
  children: React.ReactNode;
  compact?: boolean;
  description?: string;
  actions?: React.ReactNode;
};

export function Section({ title, children, compact = false, description, actions }: SectionProps) {
  return (
    <section className={compact ? 'panel panel--compact' : 'panel'}>
      <div className="panel__header">
        <div className="panel__heading">
          <h2>{title}</h2>
          {description ? <p>{description}</p> : null}
        </div>
        {actions ? <div className="panel__actions">{actions}</div> : null}
      </div>
      {children}
    </section>
  );
}

export function DataTable({ children }: { children: React.ReactNode }) {
  return <div className="table-wrap">{children}</div>;
}