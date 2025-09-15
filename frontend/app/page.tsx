"use client"

import Image from "next/image"
import { useSession, signIn, signOut } from "next-auth/react"
import { jwtDecode } from "jwt-decode"
import { useEffect, useState } from "react"

export default function Home() {
  const { data: session } = useSession()
  const [missions, setMissions] = useState<any[]>([])
  const [apiError, setApiError] = useState<string | null>(null)

  // Decode JWT to check roles
  let decodedToken: any = null
  if (session && (session as any).accessToken) {
    try {
      decodedToken = jwtDecode((session as any).accessToken)
    } catch (err) {
      console.error("Failed to decode token", err)
    }
  }

  const roles =
    decodedToken?.realm_access?.roles || ([] as string[])
  const isAdmin = roles.includes("admin")

  // Fetch missions
  useEffect(() => {
    const fetchMissions = async () => {
      if (!session) return
      try {
        const res = await fetch("http://localhost:8080/missions", {
          headers: {
            Authorization: `Bearer ${(session as any).accessToken}`,
          },
        })
        if (!res.ok) {
          throw new Error(await res.text())
        }
        const data = await res.json()
        setMissions(data)
      } catch (err: any) {
        setApiError(err.message || "Failed to fetch missions")
      }
    }
    fetchMissions()
  }, [session])

  // Handlers for admin actions
  const addMission = async () => {
    const res = await fetch("http://localhost:8080/missions", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${(session as any).accessToken}`,
      },
      body: JSON.stringify({ name: "New Mission", status: "planned" }),
    })
    const newMission = await res.json()
    setMissions((prev) => [...prev, newMission])
  }

  const deleteMission = async (id: string) => {
    await fetch(`http://localhost:8080/missions/${id}`, {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${(session as any).accessToken}`,
      },
    })
    setMissions((prev) => prev.filter((m) => m.id !== id))
  }

  return (
    <div className="font-sans min-h-screen p-8 bg-black text-white">
      {!session ? (
        <div className="flex flex-col items-center">
          <p className="text-red-400">Not signed in</p>
          <button
            onClick={() => signIn("keycloak")}
            className="mt-2 px-4 py-2 bg-blue-600 text-white rounded"
          >
            Sign in with Keycloak
          </button>
        </div>
      ) : (
        <>
          <h1 className="text-xl font-bold">
            Welcome {session.user?.email}
          </h1>

          {/* Missions */}
          <section className="mt-6">
            <h2 className="text-lg font-semibold">Missions</h2>
            {apiError ? (
              <p className="text-red-400">API Error: {apiError}</p>
            ) : (
              <ul className="list-disc ml-6">
                {missions.map((m) => (
                  <li key={m.id}>
                    {m.name} — <i>{m.status}</i>
                    {isAdmin && (
                      <button
                        onClick={() => deleteMission(m.id)}
                        className="ml-2 px-2 py-1 text-xs bg-red-600 rounded"
                      >
                        Delete
                      </button>
                    )}
                  </li>
                ))}
              </ul>
            )}

            {isAdmin && (
              <button
                onClick={addMission}
                className="mt-4 px-4 py-2 bg-green-600 rounded"
              >
                ➕ Add Mission
              </button>
            )}
          </section>

          {/* Debug: decoded JWT */}
          {decodedToken && (
            <section className="mt-6 p-4 bg-gray-900 rounded text-xs overflow-x-auto">
              <h3 className="font-bold mb-2">Decoded JWT:</h3>
              <pre>{JSON.stringify(decodedToken, null, 2)}</pre>
            </section>
          )}

          <button
            onClick={() => signOut()}
            className="mt-6 px-4 py-2 bg-red-600 rounded"
          >
            Sign out
          </button>
        </>
      )}
    </div>
  )
}
