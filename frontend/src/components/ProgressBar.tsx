export function ProgressBar() {
  return (
    <div role="status" aria-label="Loading" className="w-full h-0.5 bg-[var(--color-card-10)] overflow-hidden">
      <div className="h-full w-1/3 bg-gradient-to-r from-[var(--color-accent-lt)] to-[var(--color-accent)] animate-progress" />
    </div>
  )
}
