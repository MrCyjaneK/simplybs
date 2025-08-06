<!DOCTYPE html>
<html>
<head>
    <title>{{.Package.Package}} - Package Details</title>
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
        
        .download-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-top: 20px;
        }
        
        .download-card {
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            padding: 20px;
            border-left: 4px solid #007bff;
        }
        
        .download-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 15px;
        }
        
        .download-title {
            font-weight: 600;
            color: #2c3e50;
            font-size: 1.1em;
        }
        
        .download-builder {
            background: #e9ecef;
            color: #495057;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 0.85em;
            font-weight: 500;
        }
        
        .download-details {
            display: flex;
            flex-direction: column;
            gap: 8px;
            margin-bottom: 15px;
        }
        
        .download-detail {
            display: flex;
            justify-content: space-between;
            font-size: 0.9em;
        }
        
        .download-detail-label {
            color: #6c757d;
            font-weight: 500;
        }
        
        .download-detail-value {
            color: #495057;
            font-family: 'SF Mono', Consolas, monospace;
        }
        
        .download-buttons {
            display: flex;
            gap: 10px;
        }
        
        .download-btn {
            flex: 1;
            padding: 8px 16px;
            border-radius: 6px;
            text-decoration: none;
            text-align: center;
            font-weight: 500;
            font-size: 0.9em;
            transition: all 0.2s;
        }
        
        .download-btn-primary {
            background: #007bff;
            color: white;
        }
        
        .download-btn-primary:hover {
            background: #0056b3;
            color: white;
        }
        
        .download-btn-secondary {
            background: #6c757d;
            color: white;
        }
        
        .download-btn-secondary:hover {
            background: #545b62;
            color: white;
        }
        
        .build-matrix {
            margin-top: 20px;
            overflow-x: auto;
        }
        
        .matrix-table {
            border-collapse: collapse;
            width: 100%;
            min-width: 800px;
            background: white;
            border-radius: 6px;
            overflow: hidden;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        
        .matrix-table th {
            background-color: #f8f9fa;
            padding: 12px 8px;
            text-align: center;
            border: 1px solid #e9ecef;
            font-weight: 600;
            color: #495057;
            font-size: 0.9em;
        }
        
        .matrix-table td {
            padding: 8px;
            text-align: center;
            border: 1px solid #e9ecef;
            font-size: 0.85em;
        }
        
        .matrix-table .target-label {
            background-color: #f8f9fa;
            font-weight: 500;
            text-align: left;
            padding-left: 12px;
            color: #495057;
        }
        
        .matrix-cell {
            display: block;
            padding: 8px 4px;
            border-radius: 4px;
            text-decoration: none;
            font-weight: 500;
            transition: all 0.2s;
            min-height: 16px;
        }
        
        .matrix-cell-available {
            background-color: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        
        .matrix-cell-available:hover {
            background-color: #c3e6cb;
            color: #155724;
            text-decoration: none;
            transform: scale(1.05);
        }
        
        .matrix-cell-unavailable {
            background-color: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
            cursor: not-allowed;
        }
        
        .matrix-cell-unavailable:hover {
            text-decoration: none;
            color: #721c24;
        }
        
        /* Smooth scrolling for anchor links */
        html {
            scroll-behavior: smooth;
        }
        
        /* Highlight target when navigated to via anchor */
        .download-card:target {
            animation: highlight 2s ease-in-out;
        }
        
        @keyframes highlight {
            0% { 
                background-color: #fff3cd; 
                transform: scale(1.02);
                box-shadow: 0 4px 20px rgba(255, 193, 7, 0.3);
            }
            100% { 
                background-color: white; 
                transform: scale(1);
                box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            }
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
          <a href="{{getRelativePath .Package.Package "index"}}" class="back-link">← Back to Package List</a>
          
          <h1>{{.Package.Package}}</h1>
    
    <div class="info-section">
        <h2>Basic Information</h2>
        <table class="info-table">
            <tr><th>Package Name</th><td>{{.Package.Package}}</td></tr>
            <tr><th>Version</th><td>{{.Package.Version}}</td></tr>
            <tr><th>Type</th><td>{{.Package.Type}}</td></tr>
        </table>
    </div>

    <div class="info-section">
        <h2>Source Download</h2>
        <table class="info-table">
            <tr><th>Kind</th><td>{{.Package.Download.Kind}}</td></tr>
            <tr><th>URL</th><td><a href="{{getMirrorPath .}}" download>{{.Package.Download.URL}}</a></td></tr>
            <tr><th>SHA256</th><td><code>{{.Package.Download.Sha256}}</code></td></tr>
        </table>
    </div>

    {{if .Package.Dependencies}}
    <div class="info-section">
        <h2>Dependencies Explorer</h2>
        <div class="content-list">
            {{range .Package.Dependencies}}
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
                            <a href="{{getRelativePath $.Package.Package $dep}}" class="dependency-link">{{$dep}}</a>
                        {{else}}
                            <span class="dependency-missing">{{$dep}}</span>
                        {{end}}
                    </div>
                </div>
            {{end}}
        </div>
    </div>
    {{end}}

    {{if .Package.Build.Env}}
    <div class="info-section">
        <h2>Build Environment</h2>
        <div class="content-list">
            {{range .Package.Build.Env}}
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

    {{if .Package.Build.Steps}}
    <div class="info-section">
        <h2>Build Steps</h2>
        <div class="content-list">
            {{range $index, $step := .Package.Build.Steps}}
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

    <div class="info-section">
        <h2>Build Matrix</h2>
        <div class="build-matrix">
            <table class="matrix-table">
                <thead>
                    <tr>
                        <th>Target / Builder</th>
                        {{range getBuilders}}
                        <th>{{.}}</th>
                        {{end}}
                    </tr>
                </thead>
                <tbody>
                    {{$matrix := getBuildMatrix .}}
                    {{range getTargets}}
                    {{$target := .}}
                    <tr>
                        <td class="target-label">{{$target}}</td>
                        {{range getBuilders}}
                        {{$builder := .}}
                        {{$build := index (index $matrix $builder) $target}}
                        <td>
                            {{if $build}}
                                <a href="#download-{{$builder}}-{{$target}}" 
                                   class="matrix-cell matrix-cell-available" 
                                   title="View {{$builder}} build details for {{$target}} ({{formatFileSize $build.FileSize}})">
                                    ✓
                                </a>
                            {{else}}
                                <span class="matrix-cell matrix-cell-unavailable" 
                                      title="No {{$builder}} build available for {{$target}}">
                                    ✗
                                </span>
                            {{end}}
                        </td>
                        {{end}}
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
    </div>

    {{if .BuiltFiles}}
    <div class="info-section">
        <h2>Available Downloads</h2>
        <div class="download-grid">
            {{range .BuiltFiles}}
            <div class="download-card" id="download-{{.Builder}}-{{.Target}}">
                <div class="download-header">
                    <div class="download-title">{{.Target}}</div>
                    <div class="download-builder">{{.Builder}}</div>
                </div>
                <div class="download-details">
                    <div class="download-detail">
                        <span class="download-detail-label">Size:</span>
                        <span class="download-detail-value">{{formatFileSize .FileSize}}</span>
                    </div>
                    <div class="download-detail">
                        <span class="download-detail-label">Build ID:</span>
                        <span class="download-detail-value">{{.ID}}</span>
                    </div>
                </div>
                <div class="download-buttons">
                    <a href="{{getBuiltFilePath $.Package.Package .ArchPath}}" class="download-btn download-btn-primary" download>
                        ⬇ Tar.gz
                    </a>
                    <a href="{{getBuiltFilePath $.Package.Package .InfoPath}}" class="download-btn download-btn-secondary" target="_blank">
                        📄 Info
                    </a>
                </div>
            </div>
            {{end}}
        </div>
    </div>
    {{end}}
    
      </div> <!-- container -->
  </body>
  </html>