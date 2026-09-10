export function StatusPill({ value }: { value: string }) {
  const tone = value.toLowerCase();
  const className =
    tone.includes('success') || tone.includes('healthy') || tone.includes('ready') || tone.includes('approved') || tone.includes('connected')
      ? 'status-pill status-pill--good'
      : tone.includes('running') || tone.includes('busy') || tone.includes('queued')
        ? 'status-pill status-pill--warn'
        : tone.includes('degraded') || tone.includes('offline') || tone.includes('rejected')
          ? 'status-pill status-pill--bad'
          : 'status-pill';

  return <span className={className}>{value}</span>;
}