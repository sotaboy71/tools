package com.pentoolkit.app;

import android.annotation.SuppressLint;
import android.os.Bundle;
import android.view.KeyEvent;
import android.webkit.WebResourceError;
import android.webkit.WebResourceRequest;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;

import androidx.appcompat.app.AppCompatActivity;

/**
 * A thin native wrapper that displays the pentoolkit web UI in a WebView.
 *
 * The UI itself is served by the pentoolkit binary's "serve" command, which
 * you run locally on the device (for example inside Termux):
 *
 *     ./pentoolkit serve
 *
 * This activity simply points a WebView at that local server. If the server is
 * not running, a short "how to start it" page is shown with a Retry button.
 */
public class MainActivity extends AppCompatActivity {

    /**
     * Where the pentoolkit server is listening. Change the port here if you
     * start the server with a different -addr.
     */
    private static final String SERVER_URL = "http://127.0.0.1:8787/";

    private WebView webView;

    @SuppressLint("SetJavaScriptEnabled")
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        webView = new WebView(this);
        setContentView(webView);

        WebSettings s = webView.getSettings();
        s.setJavaScriptEnabled(true);       // the UI is a single-page JS app
        s.setDomStorageEnabled(true);
        s.setLoadWithOverviewMode(true);
        s.setUseWideViewPort(true);

        webView.setWebViewClient(new WebViewClient() {
            @Override
            public boolean shouldOverrideUrlLoading(WebView view, WebResourceRequest request) {
                // Keep navigation inside the WebView.
                view.loadUrl(request.getUrl().toString());
                return true;
            }

            @Override
            public void onReceivedError(WebView view, WebResourceRequest request, WebResourceError error) {
                // Only replace the page for the top-level request, not sub-resources.
                if (request.isForMainFrame()) {
                    view.loadDataWithBaseURL(null, notRunningHtml(), "text/html", "utf-8", null);
                }
            }
        });

        if (savedInstanceState != null) {
            webView.restoreState(savedInstanceState);
        } else {
            webView.loadUrl(SERVER_URL);
        }
    }

    @Override
    protected void onSaveInstanceState(Bundle outState) {
        super.onSaveInstanceState(outState);
        webView.saveState(outState);
    }

    @Override
    public boolean onKeyDown(int keyCode, KeyEvent event) {
        // Use the hardware/gesture Back button to navigate WebView history.
        if (keyCode == KeyEvent.KEYCODE_BACK && webView.canGoBack()) {
            webView.goBack();
            return true;
        }
        return super.onKeyDown(keyCode, event);
    }

    /** A friendly fallback page shown when the local server is unreachable. */
    private String notRunningHtml() {
        return "<!DOCTYPE html><html><head><meta name='viewport' "
                + "content='width=device-width, initial-scale=1'>"
                + "<style>body{font-family:sans-serif;background:#0b1020;color:#e7ecff;"
                + "margin:0;padding:24px;line-height:1.6}h1{font-size:20px}code{background:"
                + "#1d2540;padding:2px 6px;border-radius:6px}pre{background:#080c18;border:1px"
                + " solid #2a3358;border-radius:10px;padding:12px;overflow:auto}"
                + "button{margin-top:20px;padding:14px 20px;font-size:16px;font-weight:700;"
                + "color:#fff;background:#2b6bff;border:none;border-radius:12px}</style></head>"
                + "<body><h1>Can’t reach the pentoolkit server</h1>"
                + "<p>Start it on this device (for example in Termux), then tap Retry:</p>"
                + "<pre>./pentoolkit serve</pre>"
                + "<p>It listens on <code>" + SERVER_URL + "</code> by default. "
                + "If you used a different port, update <code>SERVER_URL</code> in "
                + "<code>MainActivity.java</code> and rebuild.</p>"
                + "<button onclick=\"location.href='" + SERVER_URL + "'\">Retry</button>"
                + "</body></html>";
    }
}
