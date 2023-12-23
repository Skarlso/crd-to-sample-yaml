const cacheName = "app-" + "4b705cbe38f400c5db0320ed10fc29357d6d1889";
const resourcesToCache = ["/","/app.css","/app.js","/manifest.webmanifest","/wasm_exec.js","/web/app.wasm","/web/css/alert.css","/web/img/logo.png","https://cdn.jsdelivr.net/npm/bootstrap@5.3.2/dist/js/bootstrap.bundle.min.js","https://cdn.jsdelivr.net/npm/halfmoon@2.0.1/css/halfmoon.min.css","https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css","https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js","https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/languages/yaml.min.js","https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/default.min.css"];

self.addEventListener("install", (event) => {
  console.log("installing app worker 4b705cbe38f400c5db0320ed10fc29357d6d1889");

  event.waitUntil(
    caches
      .open(cacheName)
      .then((cache) => {
        return cache.addAll(resourcesToCache);
      })
      .then(() => {
        self.skipWaiting();
      })
  );
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches.keys().then((keyList) => {
      return Promise.all(
        keyList.map((key) => {
          if (key !== cacheName) {
            return caches.delete(key);
          }
        })
      );
    })
  );
  console.log("app worker 4b705cbe38f400c5db0320ed10fc29357d6d1889 is activated");
});

self.addEventListener("fetch", (event) => {
  event.respondWith(
    caches.match(event.request).then((response) => {
      return response || fetch(event.request);
    })
  );
});

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
