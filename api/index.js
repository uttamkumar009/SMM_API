function setHeaders(res, req) {
  const origin = req.headers.origin;
  if (origin && (origin.startsWith("https://") || origin === "http://localhost:5500")) {
    res.setHeader("Access-Control-Allow-Origin", origin);
    res.setHeader("Vary", "Origin");
  }
  res.setHeader("Access-Control-Allow-Headers", "Authorization, Content-Type");
  res.setHeader("Access-Control-Allow-Methods", "GET, POST, OPTIONS");
  res.setHeader("X-Content-Type-Options", "nosniff");
}

const demoUser = {
  id: "usr_demo",
  name: "Alex Kumar",
  email: "demo@orbitpanel.com",
  balance: 248.90
};

module.exports = (req, res) => {
  setHeaders(res, req);
  if (req.method === "OPTIONS") {
    res.status(204).end();
    return;
  }

  const path = req.url.split("?")[0];
  if (path === "/api/health" && req.method === "GET") {
    res.status(200).json({ ok: true, message: "smm-api-vercel" });
    return;
  }
  if (path === "/api/auth/login" && req.method === "POST") {
    const body = req.body || {};
    if (body.email !== "demo@orbitpanel.com" || body.password !== "demo123") {
      res.status(401).json({ error: "invalid email or password" });
      return;
    }
    res.status(200).json({
      token: `vercel-demo-token-${Date.now()}`,
      user: demoUser
    });
    return;
  }
  if (path === "/api/me" && req.method === "GET") {
    res.status(200).json({ user: demoUser });
    return;
  }
  if (path === "/api/orders" && req.method === "GET") {
    res.status(200).json({ orders: [] });
    return;
  }
  res.status(404).json({ error: "route not found" });
};
