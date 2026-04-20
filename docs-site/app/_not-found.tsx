import Link from "next/link";

export default function NotFound() {
  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        minHeight: "100vh",
        padding: "2rem",
      }}
    >
      <h1 style={{ fontSize: "3rem", marginBottom: "1rem" }}>404</h1>
      <p style={{ fontSize: "1.25rem", marginBottom: "2rem" }}>
        페이지를 찾을 수 없습니다
      </p>
      <Link
        href="/"
        style={{
          padding: "0.75rem 1.5rem",
          backgroundColor: "#0070f3",
          color: "white",
          borderRadius: "0.375rem",
          textDecoration: "none",
        }}
      >
        홈으로 이동
      </Link>
    </div>
  );
}
