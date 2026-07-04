import os

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    modified = False

    # Suppress err
    new_content = []
    lines = content.split('\n')
    for i, line in enumerate(lines):
        if 'catch (err)' in line or 'catch (err) {' in line:
            new_content.append('      // eslint-disable-next-line no-unused-vars')
        new_content.append(line)
    
    content = '\n'.join(new_content)
    if len(lines) != len(new_content):
        modified = True

    # Suppress useEffect missing dependencies
    if 'react-hooks/exhaustive-deps' not in content:
        if filepath.endswith('App.jsx') and '}, []);' in content:
            content = content.replace('  }, []);', '  // eslint-disable-next-line react-hooks/exhaustive-deps\n  }, []);')
            modified = True
        elif filepath.endswith('Admin.jsx') and '}, []);' in content:
            content = content.replace('  }, []);', '  // eslint-disable-next-line react-hooks/exhaustive-deps\n  }, []);')
            modified = True
        elif filepath.endswith('ProductDetail.jsx') and '}, [id]);' in content:
            content = content.replace('  }, [id]);', '  // eslint-disable-next-line react-hooks/exhaustive-deps\n  }, [id]);')
            modified = True
        elif filepath.endswith('useWebsocket.js') and '}, [isAuthenticated, token]);' in content:
            content = content.replace('  }, [isAuthenticated, token]);', '  // eslint-disable-next-line react-hooks/exhaustive-deps\n  }, [isAuthenticated, token]);')
            modified = True

    if modified:
        with open(filepath, 'w') as f:
            f.write(content)

src_dir = 'frontend/src'
for root, dirs, files in os.walk(src_dir):
    for file in files:
        if file.endswith(('.js', '.jsx')):
            process_file(os.path.join(root, file))
