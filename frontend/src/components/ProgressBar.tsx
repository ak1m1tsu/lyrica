export function ProgressBar() {
  return (
    <div role="status" aria-label="Loading" className="w-full h-0.5 bg-gray-200 dark:bg-white/10 overflow-hidden">
      <div className="h-full w-1/3 bg-gray-400 dark:bg-white/50 animate-progress" />
    </div>
  )
}
