export function ProgressBar() {
  return (
    <div role="status" aria-label="Loading" className="w-full h-0.5 bg-[#0f1117]/10 dark:bg-white/10 overflow-hidden">
      <div className="h-full w-1/3 bg-gradient-to-r from-[#C8B1F3] to-[#9B84D1] animate-progress" />
    </div>
  )
}
