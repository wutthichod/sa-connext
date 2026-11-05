// app/page.tsx
import { redirect } from 'next/navigation'

export default function Page() {
  // server-side redirect — no flash, instant
  redirect('/login')
  return null
}
