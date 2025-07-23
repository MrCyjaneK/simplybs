<!DOCTYPE html>
<html>
<head>
    <title>{{.Package}} - Package Details</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; 
            margin: 0; 
            padding: 40px;
            background-color: #f8f9fa;
            color: #333;
        }
        .container { 
            max-width: 1200px; 
            margin: 0 auto; 
            background: white; 
            border-radius: 8px; 
            box-shadow: 0 2px 10px rgba(0,0,0,0.1); 
            padding: 30px; 
        }
        h1 { color: #2c3e50; margin-bottom: 10px; font-size: 2.5em; }
        h2 { color: #34495e; border-bottom: 2px solid #e9ecef; padding-bottom: 10px; margin-top: 30px; }
        .back-link { 
            margin-bottom: 20px; 
            display: inline-block;
            padding: 8px 16px;
            background: #6c757d;
            color: white;
            border-radius: 5px;
            text-decoration: none;
        }
        .back-link:hover { background: #545b62; color: white; }
        .info-section { margin: 25px 0; }
        .info-table { 
            border-collapse: collapse; 
            width: 100%; 
            background: white;
            border-radius: 6px;
            overflow: hidden;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        .info-table th, .info-table td { 
            padding: 12px 16px; 
            text-align: left; 
            border-bottom: 1px solid #e9ecef;
        }
        .info-table th { 
            background-color: #f8f9fa; 
            font-weight: 600;
            width: 180px; 
            color: #495057;
        }
        .info-table tr:last-child td { border-bottom: none; }
        
        .content-list {
            background: white;
            border-radius: 6px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .content-item {
            display: flex;
            align-items: center;
            padding: 12px 16px;
            border-bottom: 1px solid #e9ecef;
            min-height: 24px;
        }
        .content-item:last-child { border-bottom: none; }
        .content-item:nth-child(even) { background-color: #f8f9fa; }
        
        .glob-column {
            width: 150px;
            flex-shrink: 0;
            margin-right: 16px;
        }
        .glob-pattern { 
            display: inline-block;
            background: #e9ecef;
            color: #495057;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 0.85em;
            font-weight: 500;
            min-width: 60px;
            text-align: center;
        }
        .glob-empty {
            color: #adb5bd;
            font-style: italic;
        }
        
        .content-column {
            flex: 1;
            min-width: 0;
        }
        .content-text {
            word-break: break-word;
            font-family: 'SF Mono', Consolas, monospace;
            font-size: 0.9em;
        }
        
        .dependency-link { 
            color: #007bff; 
            text-decoration: none; 
            font-weight: 500;
            padding: 2px 6px;
            border-radius: 3px;
            transition: background-color 0.2s;
        }
        .dependency-link:hover { 
            background-color: #e3f2fd;
            text-decoration: none; 
        }
        .dependency-missing { 
            color: #6c757d; 
            font-style: italic;
        }
        
        .step-number {
            width: 30px;
            flex-shrink: 0;
            margin-right: 12px;
            font-weight: bold;
            color: #495057;
        }
        
        a { color: #007bff; text-decoration: none; }
        a:hover { text-decoration: underline; }
        
        code { 
            background: #f8f9fa; 
            padding: 2px 4px; 
            border-radius: 3px; 
            font-family: 'SF Mono', Consolas, monospace;
            font-size: 0.9em;
        }
    </style>
  </head>
  <body>
      <div class="container">
          <a href="index.html" class="back-link">← Back to Package List</a>
          
          <h1>{{.Package}}</h1>
    
    <div class="info-section">
        <h2>Basic Information</h2>
        <table class="info-table">
            <tr><th>Package Name</th><td>{{.Package}}</td></tr>
            <tr><th>Version</th><td>{{.Version}}</td></tr>
            <tr><th>Type</th><td>{{.Type}}</td></tr>
        </table>
    </div>

    <div class="info-section">
        <h2>Download Information</h2>
        <table class="info-table">
            <tr><th>Kind</th><td>{{.Download.Kind}}</td></tr>
            <tr><th>URL</th><td><a href="{{.Download.URL}}" target="_blank">{{.Download.URL}}</a></td></tr>
            <tr><th>SHA256</th><td><code>{{.Download.Sha256}}</code></td></tr>
        </table>
    </div>

    {{if .Dependencies}}
    <div class="info-section">
        <h2>Dependencies Explorer</h2>
        <div class="content-list">
            {{range .Dependencies}}
                {{$pattern := getGlobPattern .}}
                {{$dep := getGlobContent .}}
                <div class="content-item">
                    <div class="glob-column">
                        {{if $pattern}}
                            <span class="glob-pattern">{{$pattern}}</span>
                        {{else}}
                            <span class="glob-pattern glob-empty">all</span>
                        {{end}}
                    </div>
                    <div class="content-column">
                        {{if depExists $dep}}
                            <a href="{{$dep}}.html" class="dependency-link">{{$dep}}</a>
                        {{else}}
                            <span class="dependency-missing">{{$dep}}</span>
                        {{end}}
                    </div>
                </div>
            {{end}}
        </div>
    </div>
    {{end}}

    {{if .Build.Env}}
    <div class="info-section">
        <h2>Build Environment</h2>
        <div class="content-list">
            {{range .Build.Env}}
                {{$pattern := getGlobPattern .}}
                {{$content := getGlobContent .}}
                <div class="content-item">
                    <div class="glob-column">
                        {{if $pattern}}
                            <span class="glob-pattern">{{$pattern}}</span>
                        {{else}}
                            <span class="glob-pattern glob-empty">all</span>
                        {{end}}
                    </div>
                    <div class="content-column">
                        <code class="content-text">{{$content}}</code>
                    </div>
                </div>
            {{end}}
        </div>
    </div>
    {{end}}

    {{if .Build.Steps}}
    <div class="info-section">
        <h2>Build Steps</h2>
        <div class="content-list">
            {{range $index, $step := .Build.Steps}}
                {{$pattern := getGlobPattern $step}}
                {{$content := getGlobContent $step}}
                <div class="content-item">
                    <div class="step-number">{{add $index 1}}.</div>
                    <div class="glob-column">
                        {{if $pattern}}
                            <span class="glob-pattern">{{$pattern}}</span>
                        {{else}}
                            <span class="glob-pattern glob-empty">all</span>
                        {{end}}
                    </div>
                    <div class="content-column">
                        <code class="content-text">{{$content}}</code>
                    </div>
                </div>
            {{end}}
        </div>
    </div>
    {{end}}
    
      </div> <!-- container -->
  </body>
  </html>