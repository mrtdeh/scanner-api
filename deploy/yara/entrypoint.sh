#!/bin/sh

echo "Starting YARA Service..."
check_yara_rules() {
    local target_dir="${1:-/rules}"
    
    if [ ! -d "$target_dir" ]; then
        echo "ERROR: Directory $target_dir does not exist!"
        return 1
    fi
    
    yar_files=$(find "$target_dir" -maxdepth 1 -type f -name "*.yar" 2>/dev/null)
    
    if [ -z "$yar_files" ]; then
        echo "ERROR: No .yar files found in $target_dir"
        return 1
    fi
        
    return 0
}



if ! check_yara_rules "$1"; then
    echo "Failed"
    exit 1
fi

while true; do
    sleep 1
done

# ==========================================


