// -----------------------------------------------------------------------------
// PWA
// -----------------------------------------------------------------------------
const cacheName = "app-" + "v1.0.0";
const resourcesToCache = ["https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/themes/prism-tomorrow.min.css","https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/autoloader/prism-autoloader.min.js","https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-core.min.js","https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css","https://cdnjs.cloudflare.com/ajax/libs/clipboard.js/2.0.11/clipboard.min.js","https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/js/bootstrap.bundle.min.js","https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css","/web/img/logo.png","/web/css/modern-style.css","/web/app.wasm","/wasm_exec.js","/manifest.webmanifest","/app.js","/app.css","/"];

self.addEventListener("install", async (event) => {
  try {
    console.log("installing app worker v1.0.0");
    await installWorker();
    await self.skipWaiting();
  } catch (error) {
    console.error("error during installation:", error);
  }
});

async function installWorker() {
  const cache = await caches.open(cacheName);
  await cache.addAll(resourcesToCache);
}

self.addEventListener("activate", async (event) => {
  try {
    await deletePreviousCaches(); // Await cache cleanup
    await self.clients.claim(); // Ensure the service worker takes control of the clients
    console.log("app worker v1.0.0 is activated");
  } catch (error) {
    console.error("error during activation:", error);
  }
});

async function deletePreviousCaches() {
  const keys = await caches.keys();
  await Promise.all(
    keys.map(async (key) => {
      if (key !== cacheName) {
        try {
          console.log("deleting", key, "cache");
          await caches.delete(key);
        } catch (err) {
          console.error("deleting", key, "cache failed:", err);
        }
      }
    })
  );
}

self.addEventListener("fetch", (event) => {
  event.respondWith(fetchWithCache(event.request));
});

async function fetchWithCache(request) {
  const cachedResponse = await caches.match(request);
  if (cachedResponse) {
    return cachedResponse;
  }
  return await fetch(request);
}

// -----------------------------------------------------------------------------
// Push Notifications
// -----------------------------------------------------------------------------
self.addEventListener("push", (event) => {
  if (!event.data || !event.data.text()) {
    return;
  }

  const notification = JSON.parse(event.data.text());
  if (!notification) {
    return;
  }

  const title = notification.title;
  delete notification.title;

  if (!notification.data) {
    notification.data = {};
  }
  let actions = [];
  for (let i in notification.actions) {
    const action = notification.actions[i];

    actions.push({
      action: action.action,
      path: action.path,
    });

    delete action.path;
  }
  notification.data.goapp = {
    path: notification.path,
    actions: actions,
  };
  delete notification.path;

  event.waitUntil(self.registration.showNotification(title, notification));
});

self.addEventListener("notificationclick", (event) => {
  event.notification.close();

  const notification = event.notification;
  let path = notification.data.goapp.path;

  for (let i in notification.data.goapp.actions) {
    const action = notification.data.goapp.actions[i];
    if (action.action === event.action) {
      path = action.path;
      break;
    }
  }

  event.waitUntil(
    clients
      .matchAll({
        type: "window",
      })
      .then((clientList) => {
        for (var i = 0; i < clientList.length; i++) {
          let client = clientList[i];
          if ("focus" in client) {
            client.focus();
            client.postMessage({
              goapp: {
                type: "notification",
                path: path,
              },
            });
            return;
          }
        }

        if (clients.openWindow) {
          return clients.openWindow(path);
        }
      })
  );
});
