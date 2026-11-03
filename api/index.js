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

async function providerRequest(values) {
  const apiUrl = process.env.SMM_PROVIDER_API_URL || "https://cheapestsmmpanels.com/api/v2";
  const apiKey = process.env.SMM_PROVIDER_API_KEY;
  if (!apiKey) throw new Error("provider API is not configured");
  values.set("key", apiKey);
  const response = await fetch(apiUrl, {
    method: "POST",
    headers: {"Content-Type": "application/x-www-form-urlencoded", "User-Agent": "BabiesiQ-SMM/1.0"},
    body: values
  });
  if (!response.ok) throw new Error(`provider returned HTTP ${response.status}`);
  return response.json();
}

module.exports = async (req, res) => {
  setHeaders(res, req);
  if (req.method === "OPTIONS") {
    res.status(204).end();
    return;
  }

  const path = req.url.split("?")[0];
  if ((path === "/api/health" || path === "/health") && req.method === "GET") {
    res.status(200).json({ ok: true, message: "smm-api-vercel" });
    return;
  }
  if ((path === "/api/services" || path === "/services") && req.method === "GET") {
    try {
      const services = await providerRequest(new URLSearchParams({action: "services"}));
      const category = new URL(req.url, "http://localhost").searchParams.get("category");
      const filtered = category ? services.filter(service => String(service.category || "").toLowerCase() === category.toLowerCase()) : services;
      res.status(200).json({services: filtered});
    } catch (error) {
      res.status(502).json({error: error.message});
    }
    return;
  }
  if ((path === "/api/categories" || path === "/categories") && req.method === "GET") {
    try {
      const services = await providerRequest(new URLSearchParams({action: "services"}));
      const categories = [...new Set(services.map(service => service.category).filter(Boolean))].sort();
      res.status(200).json({categories});
    } catch (error) {
      res.status(502).json({error: error.message});
    }
    return;
  }
  if ((path === "/api/auth/login" || path === "/auth/login") && req.method === "POST") {
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
  if ((path === "/api/me" || path === "/me") && req.method === "GET") {
    res.status(200).json({ user: demoUser });
    return;
  }
  if ((path === "/api/orders" || path === "/orders") && req.method === "GET") {
    res.status(200).json({ orders: [] });
    return;
  }
  if ((path === "/api/orders" || path === "/orders") && req.method === "POST") {
    const body = req.body || {};
    if (!body.serviceId || !body.link || !body.quantity) {
      res.status(400).json({error: "serviceId, link and positive quantity are required"});
      return;
    }
    try {
      const providerOrder = await providerRequest(new URLSearchParams({
        action: "add",
        service: String(body.serviceId),
        link: String(body.link),
        quantity: String(body.quantity)
      }));
      if (providerOrder.error) {
        res.status(502).json({error: providerOrder.error});
        return;
      }
      res.status(201).json({order: {
        id: `#OR-${Date.now().toString(36).toUpperCase()}`,
        providerId: String(providerOrder.order || ""),
        service: body.service || "SMM service",
        link: body.link,
        quantity: Number(body.quantity),
        status: "processing",
        amount: Number(body.quantity) / 1000 * Number(body.rate || 2.4)
      }});
    } catch (error) {
      res.status(502).json({error: error.message});
    }
    return;
  }
  res.status(404).json({ error: "route not found" });
};
