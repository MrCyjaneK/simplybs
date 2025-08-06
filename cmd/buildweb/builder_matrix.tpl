<!DOCTYPE html>
<html>
<head>
    <title>{{.Builder}} Build Matrix</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; 
            margin: 0; 
            padding: 40px;
            background-color: #f8f9fa;
            color: #333;
        }
        .container { 
            max-width: 1400px; 
            margin: 0 auto; 
            background: white; 
            border-radius: 8px; 
            box-shadow: 0 2px 10px rgba(0,0,0,0.1); 
            padding: 30px; 
        }
        h1 { color: #2c3e50; margin-bottom: 10px; font-size: 2.5em; }
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
        .stats { color: #6c757d; font-size: 1.1em; margin-bottom: 30px; }
        .matrix-container {
            overflow-x: auto;
            margin-top: 20px;
        }
        .matrix-table {
            border-collapse: collapse;
            width: 100%;
            min-width: 1200px;
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
            position: sticky;
            top: 0;
            z-index: 10;
        }
        .matrix-table th:first-child {
            text-align: left;
            padding-left: 16px;
            min-width: 200px;
            position: sticky;
            left: 0;
            z-index: 11;
            background-color: #f8f9fa;
        }
        .matrix-table td {
            padding: 8px;
            text-align: center;
            border: 1px solid #e9ecef;
            font-size: 0.85em;
        }
        .matrix-table .package-label {
            background-color: #f8f9fa;
            font-weight: 500;
            text-align: left;
            padding-left: 16px;
            color: #495057;
            position: sticky;
            left: 0;
            z-index: 9;
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
        .package-link {
            color: #007bff;
            text-decoration: none;
            font-weight: 500;
        }
        .package-link:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <a href="index.html" class="back-link">← Back to Package List</a>
        
        <h1>{{.Builder}} Build Matrix</h1>
        <p class="stats">Builder: <strong>{{.Builder}}</strong> | Total packages: {{len .Packages}}</p>
        
        <div class="matrix-container">
            <table class="matrix-table">
                <thead>
                    <tr>
                        <th>Package</th>
                        {{range getTargets}}
                        <th>{{.}}</th>
                        {{end}}
                    </tr>
                </thead>
                <tbody>
                    {{range .Packages}}
                    {{$pkg := .}}
                    <tr>
                        <td class="package-label">
                            <a href="{{.Package.Package}}.html" class="package-link">{{.Package.Package}}</a>
                        </td>
                        {{range getTargets}}
                        {{$target := .}}
                        {{$found := false}}
                        {{range $pkg.BuiltFiles}}
                        {{if and (eq .Builder $.Builder) (eq .Target $target)}}
                        {{$found = true}}
                        <td>
                            <a href="files/{{.Builder}}/{{.Target}}/{{$pkg.Package.Package}}-{{$pkg.Package.Version}}-{{.ID}}.html" 
                               class="matrix-cell matrix-cell-available" 
                               title="{{formatFileSize .FileSize}} - Build ID: {{.ID}}">
                                ✓
                            </a>
                        </td>
                        {{end}}
                        {{end}}
                        {{if not $found}}
                        <td>
                            <span class="matrix-cell matrix-cell-unavailable" 
                                  title="No build available for {{$target}}">
                                ✗
                            </span>
                        </td>
                        {{end}}
                        {{end}}
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
    </div>
</body>
</html>