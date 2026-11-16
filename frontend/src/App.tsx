import { useMemo, useState } from 'react'
import type { FormEvent } from 'react'

type ReportRow = Record<string, unknown>

type ReportResult = {
  columns: string[]
  rows: ReportRow[]
  count: number
  generatedAt: string
}

const today = new Date().toISOString().slice(0, 10)

function readableColumn(column: string): string {
  return column.replaceAll('_', ' ').toLowerCase().replace(/\b\w/g, (letter) => letter.toUpperCase())
}

function csvCell(value: unknown): string {
  const text = value === null || value === undefined ? '' : String(value)
  return `"${text.replaceAll('"', '""')}"`
}

function downloadCSV(result: ReportResult) {
  const lines = [
    result.columns.map(csvCell).join(','),
    ...result.rows.map((row) => result.columns.map((column) => csvCell(row[column])).join(',')),
  ]
  const blob = new Blob([lines.join('\n')], { type: 'text/csv;charset=utf-8' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = `charged-sales-${today}.csv`
  link.click()
  window.setTimeout(() => URL.revokeObjectURL(link.href), 0)
}

function LedgerMark() {
  return (
    <svg aria-hidden="true" className="size-5" fill="none" viewBox="0 0 24 24">
      <path d="M4 5.5h16M4 12h16M4 18.5h10" stroke="currentColor" strokeLinecap="round" strokeWidth="1.8" />
      <circle cx="17.5" cy="18.5" fill="currentColor" r="2.5" />
    </svg>
  )
}

function DownloadMark() {
  return (
    <svg aria-hidden="true" className="size-4" fill="none" viewBox="0 0 24 24">
      <path d="M12 3v11m0 0 4-4m-4 4-4-4M5 18.5h14" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" />
    </svg>
  )
}

function App() {
  const [dateFrom, setDateFrom] = useState(today)
  const [dateTo, setDateTo] = useState(today)
  const [companyID, setCompanyID] = useState('')
  const [storeCode, setStoreCode] = useState('')
  const [report, setReport] = useState<ReportResult | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const formattedGeneratedAt = useMemo(() => {
    if (!report) return ''
    return new Date(report.generatedAt).toLocaleString()
  }, [report])

  async function generateReport(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')

    const payload: Record<string, string | number> = { dateFrom, dateTo }
    if (companyID.trim()) {
      const parsedCompanyID = Number(companyID)
      if (!Number.isSafeInteger(parsedCompanyID)) {
        setError('Company ID must be a whole number.')
        return
      }
      payload.companyID = parsedCompanyID
    }
    if (storeCode.trim()) payload.storeCode = storeCode.trim()

    setLoading(true)
    setReport(null)

    try {
      const response = await fetch('/api/reports/charged-sales', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      })
      const body = await response.json() as ReportResult | { message?: string }
      if (!response.ok) {
        throw new Error('message' in body ? body.message : 'The report could not be generated.')
      }
      setReport(body as ReportResult)
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : 'The report could not be generated.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <main className="min-h-screen bg-canvas text-ink selection:bg-signal/25">
      <header className="border-b border-ink/10 bg-paper">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-5 py-4 sm:px-8">
          <div className="flex items-center gap-3">
            <span className="grid size-10 place-items-center rounded-[14px] bg-ink text-paper shadow-[0_8px_20px_rgba(16,42,44,0.18)]">
              <LedgerMark />
            </span>
            <div>
              <p className="font-display text-lg leading-none tracking-[-0.025em]">AR exception desk</p>
              <p className="mt-1 text-xs text-muted">Personal Oracle reporting tool</p>
            </div>
          </div>
          <span className="hidden items-center gap-2 text-sm text-muted sm:flex">
            <span className="size-2 rounded-full bg-moss" />
            Read-only source
          </span>
        </div>
      </header>

      <div className="mx-auto max-w-7xl px-5 py-10 sm:px-8 sm:py-14">
        <section className="grid gap-8 border-b border-ink/10 pb-10 lg:grid-cols-[minmax(0,1.25fr)_minmax(260px,.75fr)] lg:items-end">
          <div className="max-w-3xl">
            <p className="mb-4 text-sm font-semibold text-moss">Receivables exception monitor</p>
            <h1 className="font-display text-5xl leading-[0.95] tracking-[-0.055em] text-ink sm:text-6xl">Find sales that need an AR follow-up.</h1>
            <p className="mt-6 max-w-2xl text-base leading-7 text-muted sm:text-lg">Run a focused Charged Sales check against Oracle, review the exception rows, then take the current result with you as a CSV.</p>
          </div>
          <aside className="border-l-2 border-signal pl-5 text-sm leading-6 text-muted">
            The report checks eligible sales and returns the records that do not yet have a matching AR transaction.
          </aside>
        </section>

        <section className="mt-8 border border-ink/10 bg-paper shadow-[0_18px_50px_rgba(16,42,44,0.08)]" aria-labelledby="report-filters">
          <div className="flex flex-col gap-3 border-b border-ink/10 px-5 py-5 sm:flex-row sm:items-center sm:justify-between sm:px-7">
            <div>
              <h2 className="font-display text-2xl tracking-[-0.03em]" id="report-filters">Build a report</h2>
              <p className="mt-1 text-sm text-muted">Dates are required. Company and store narrow the result when needed.</p>
            </div>
            <p className="text-sm text-muted">Oracle data is never changed.</p>
          </div>

          <form className="grid gap-x-5 gap-y-5 px-5 py-6 sm:grid-cols-2 sm:px-7 lg:grid-cols-[1fr_1fr_1fr_1fr_auto] lg:items-end" onSubmit={generateReport}>
            <label className="grid gap-2 text-sm font-semibold text-ink">
              From date
              <input className="h-11 rounded-md border border-ink/20 bg-canvas px-3 font-normal outline-none transition focus:border-moss focus:ring-4 focus:ring-moss/10" required type="date" value={dateFrom} onChange={(event) => setDateFrom(event.target.value)} />
            </label>
            <label className="grid gap-2 text-sm font-semibold text-ink">
              To date
              <input className="h-11 rounded-md border border-ink/20 bg-canvas px-3 font-normal outline-none transition focus:border-moss focus:ring-4 focus:ring-moss/10" required min={dateFrom} type="date" value={dateTo} onChange={(event) => setDateTo(event.target.value)} />
            </label>
            <label className="grid gap-2 text-sm font-semibold text-ink">
              Company ID <span className="font-normal text-muted">Optional</span>
              <input className="h-11 rounded-md border border-ink/20 bg-canvas px-3 font-normal outline-none transition placeholder:text-muted/60 focus:border-moss focus:ring-4 focus:ring-moss/10" inputMode="numeric" min="0" placeholder="For example, 10" step="1" type="number" value={companyID} onChange={(event) => setCompanyID(event.target.value)} />
            </label>
            <label className="grid gap-2 text-sm font-semibold text-ink">
              Store code <span className="font-normal text-muted">Optional</span>
              <input className="h-11 rounded-md border border-ink/20 bg-canvas px-3 font-normal outline-none transition placeholder:text-muted/60 focus:border-moss focus:ring-4 focus:ring-moss/10" maxLength={50} placeholder="For example, 001" value={storeCode} onChange={(event) => setStoreCode(event.target.value)} />
            </label>
            <button className="flex h-11 items-center justify-center gap-2 rounded-md bg-ink px-5 font-semibold text-paper transition hover:bg-moss focus:outline-none focus:ring-4 focus:ring-moss/25 disabled:cursor-not-allowed disabled:opacity-60" disabled={loading} type="submit">
              {loading && <span aria-hidden="true" className="size-4 animate-spin rounded-full border-2 border-paper/35 border-t-paper" />}
              {loading ? 'Checking Oracle' : 'Generate report'}
            </button>
          </form>
        </section>

        {error && <p className="mt-6 border-l-2 border-danger bg-danger/8 px-5 py-4 text-sm leading-6 text-danger" role="alert">{error}</p>}

        {report ? (
          <section className="mt-8 border border-ink/10 bg-paper shadow-[0_18px_50px_rgba(16,42,44,0.08)]" aria-live="polite">
            <div className="flex flex-col gap-5 border-b border-ink/10 px-5 py-6 sm:flex-row sm:items-end sm:justify-between sm:px-7">
              <div className="flex items-end gap-4">
                <span className="font-display text-5xl leading-none tracking-[-0.06em] text-moss">{report.count}</span>
                <div className="pb-1">
                  <h2 className="font-display text-2xl tracking-[-0.03em]">Charged Sales exceptions</h2>
                  <p className="mt-1 text-sm text-muted">Generated {formattedGeneratedAt}</p>
                </div>
              </div>
              <button className="inline-flex h-10 items-center justify-center gap-2 rounded-md border border-ink/20 px-4 text-sm font-semibold text-ink transition hover:border-moss hover:bg-moss/5 focus:outline-none focus:ring-4 focus:ring-moss/10 disabled:cursor-not-allowed disabled:opacity-50" disabled={report.count === 0} onClick={() => downloadCSV(report)} type="button">
                <DownloadMark />
                Download CSV
              </button>
            </div>

            {report.count === 0 ? (
              <div className="px-5 py-12 text-center sm:px-7">
                <p className="font-display text-2xl tracking-[-0.03em]">Nothing needs attention in this range.</p>
                <p className="mx-auto mt-2 max-w-md text-sm leading-6 text-muted">Try a wider period or remove a company or store filter.</p>
              </div>
            ) : (
              <div className="max-h-[34rem] overflow-auto">
                <table className="min-w-[1320px] w-full border-collapse text-left text-sm">
                  <thead className="sticky top-0 z-10 bg-ink text-paper">
                    <tr>
                      {report.columns.map((column) => <th className="px-4 py-3 text-xs font-semibold tracking-[0.01em]" key={column}>{readableColumn(column)}</th>)}
                    </tr>
                  </thead>
                  <tbody>
                    {report.rows.map((row, index) => (
                      <tr className="border-b border-ink/8 transition hover:bg-moss/5" key={`${String(row.ID)}-${index}`}>
                        {report.columns.map((column) => <td className="max-w-72 px-4 py-3 align-top text-ink/85" key={column}>{row[column] === null ? <span className="text-muted">—</span> : String(row[column] ?? '')}</td>)}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>
        ) : !loading && !error ? (
          <section className="mt-8 grid gap-4 border-t border-ink/10 pt-6 text-sm leading-6 text-muted sm:grid-cols-3">
            <p><span className="font-semibold text-ink">1.</span> Choose the period to inspect.</p>
            <p><span className="font-semibold text-ink">2.</span> Add a company or store only when useful.</p>
            <p><span className="font-semibold text-ink">3.</span> Generate a fresh, read-only result.</p>
          </section>
        ) : null}
      </div>
    </main>
  )
}

export default App
