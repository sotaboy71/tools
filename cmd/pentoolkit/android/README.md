# pentoolkit — Android WebView wrapper

A minimal Android app that wraps the pentoolkit **web UI** in a WebView, so you
get a real launcher icon and a full-screen "app" instead of a browser tab.

## How it works

This app does **not** contain the tools itself. It is a thin shell that loads
the UI served by the pentoolkit binary's `serve` command:

```
WebView  ──▶  http://127.0.0.1:8787/  ──▶  pentoolkit serve (running on-device)
```

So the workflow is:

1. Run the pentoolkit server on the phone (easiest via Termux — see
   [../TERMUX.md](../TERMUX.md)):
   ```sh
   ./pentoolkit serve
   ```
2. Open this app. It points a WebView at `http://127.0.0.1:8787/`.

If the server isn't running, the app shows a short "start it, then Retry" page.

> Why not bundle the server inside the APK? Doing that requires cross-compiling
> the Go code into an Android library with `gomobile` (or shipping the binary in
> assets and exec-ing it), which is a larger project. This wrapper keeps things
> simple and reuses the exact same UI and backend you already run via Termux. If
> you later bundle the binary, none of this UI code changes.

## Build the APK

You need the Android SDK. The easiest path is **Android Studio**:

1. Open Android Studio → *Open* → select this `android/` folder.
2. Let it sync Gradle (it will download the right Gradle and dependencies, and
   generate the Gradle wrapper if it is missing).
3. *Build → Build Bundle(s) / APK(s) → Build APK(s)*.
4. Install the resulting `app/build/outputs/apk/debug/app-debug.apk` on your
   phone (enable "install from unknown sources").

### Command line (if you have the SDK + Gradle installed)

This project does not commit the Gradle wrapper jar (a binary). Generate it
once, then build:

```sh
cd cmd/pentoolkit/android
gradle wrapper            # creates ./gradlew  (needs a local Gradle install)
./gradlew assembleDebug   # outputs app/build/outputs/apk/debug/app-debug.apk
```

Point `sdk.dir` at your Android SDK via a `local.properties` file if Gradle
can't find it:

```
sdk.dir=/path/to/Android/Sdk
```

## Configuration

- **Server address/port** — edit `SERVER_URL` in
  `app/src/main/java/com/pentoolkit/app/MainActivity.java` if you run
  `pentoolkit serve` on a different port or host.
- **Cleartext HTTP** — Android blocks plain HTTP by default; this app allows it
  **only** for `127.0.0.1`/`localhost` via
  `app/src/main/res/xml/network_security_config.xml`. If you point it at a LAN
  IP instead, add that host there too.

## Project layout

```
android/
  settings.gradle
  build.gradle
  gradle.properties
  app/
    build.gradle
    proguard-rules.pro
    src/main/
      AndroidManifest.xml
      java/com/pentoolkit/app/MainActivity.java
      res/xml/network_security_config.xml
```

## Reminder

Authorized use only — the same rules apply here as everywhere else in this
toolkit. Only scan systems you own or have explicit written permission to test.
