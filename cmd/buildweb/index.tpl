<!DOCTYPE html>
<html>
<head>
    <title>Package Information</title>
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
        .stats { color: #6c757d; font-size: 1.1em; margin-bottom: 30px; }
        table { 
            border-collapse: collapse; 
            width: 100%; 
            background: white;
            border-radius: 6px;
            overflow: hidden;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        th, td { 
            padding: 12px 16px; 
            text-align: left; 
            border-bottom: 1px solid #e9ecef;
        }
        th { 
            background-color: #f8f9fa; 
            font-weight: 600;
            color: #495057;
        }
        tr:last-child td { border-bottom: none; }
        tr:nth-child(even) { background-color: #f8f9fa; }
        .type-badge {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 0.8em;
            font-weight: 500;
        }
        .type-host { background: #e3f2fd; color: #1976d2; }
        .type-native { background: #f3e5f5; color: #7b1fa2; }
        a { 
            color: #007bff; 
            text-decoration: none; 
            padding: 6px 12px;
            border-radius: 4px;
            transition: background-color 0.2s;
        }
        a:hover { 
            background-color: #e3f2fd;
            text-decoration: none; 
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Package Information</h1>
        <p class="stats">Total packages: {{len .}}</p>
        <table>
            <tr>
                <th>Package</th>
                <th>Version</th>
                <th>Type</th>
                <th>Details</th>
            </tr>
            {{range .}}
            <tr>
                <td><strong>{{.Package}}</strong></td>
                <td>{{.Version}}</td>
                <td>
                    <span class="type-badge type-{{.Type}}">{{.Type}}</span>
                </td>
                <td><a href="{{.Package}}.html">View Details →</a></td>
            </tr>
            {{end}}
        </table>
    </div>
</body>
</html>`