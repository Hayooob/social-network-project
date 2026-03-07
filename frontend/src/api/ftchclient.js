const BASE_URL = import.meta.env.VITE_API_URL ?? (import.meta.env.DEV ? "http://localhost:8080" : "");

async function request(path, options = {}) {
  const url = `${BASE_URL}${path}`;

  const isFormData = options.body instanceof FormData;

  const config = {
    credentials: "include",
    ...options,
    headers: {
      ...(isFormData ? {} : { "Content-Type": "application/json" }),
      ...(options.headers || {}),
    },
  };

  const response = await fetch(url, config);

  let data = null;
  const contentType = response.headers.get("content-type") || "";
  if (contentType.includes("application/json")) {
    try {
      data = await response.json();
    } catch {
      data = null;
    }
  }

  if (!response.ok) {
    const message =
      data?.error ||
      data?.message ||
      `Request failed with status ${response.status}`;
    const err = new Error(message);
    err.status = response.status;
    err.data = data;
    throw err;
  }

  return data;
}

export function get(path) {
  return request(path, { method: "GET" });
}

export function post(path, body) {
  const isFormData = body instanceof FormData;

  return request(path, {
    method: "POST",
    body: body ? (isFormData ? body : JSON.stringify(body)) : undefined,
  });
}

export function del(path) {
  return request(path, { method: "DELETE" });
}

export function put(path, body) {
  const isFormData = body instanceof FormData;

  return request(path, {
    method: "PUT",
    body: body ? (isFormData ? body : JSON.stringify(body)) : undefined,
  });
}

export { request };