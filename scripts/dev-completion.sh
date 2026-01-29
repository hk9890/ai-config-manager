#!/bin/bash
# Development helper: Enable completion for ./aimgr in current shell session
#
# Usage:
#   source dev-completion.sh
#
# This enables tab completion for ./aimgr during development without installing.

if [ -n "$BASH_VERSION" ]; then
    # Bash completion
    _aimgr_dev() {
        local cur="${COMP_WORDS[COMP_CWORD]}"
        local prev="${COMP_WORDS[COMP_CWORD-1]}"
        
        # Use the __complete command to get suggestions
        local suggestions
        suggestions=$(./aimgr __complete "${COMP_WORDS[@]:1}" "${cur}" 2>/dev/null)
        
        # Parse the output (format: one suggestion per line, then :directive, then message)
        local IFS=$'\n'
        local lines=($suggestions)
        
        # Get suggestions (all lines before the :directive line)
        COMPREPLY=()
        for line in "${lines[@]}"; do
            if [[ "$line" == :* ]]; then
                break
            fi
            COMPREPLY+=("$line")
        done
    }
    
    complete -F _aimgr_dev ./aimgr
    echo "✓ Bash completion enabled for ./aimgr"
    
elif [ -n "$ZSH_VERSION" ]; then
    # Zsh completion
    _aimgr_dev() {
        local suggestions
        suggestions=(${(f)"$(./aimgr __complete ${words[@]:1} ${words[CURRENT]} 2>/dev/null)"})
        
        # Remove the :directive line and message
        suggestions=("${(@)suggestions:#:*}")
        
        compadd -a suggestions
    }
    
    compdef _aimgr_dev ./aimgr
    echo "✓ Zsh completion enabled for ./aimgr"
    
else
    echo "✗ Unknown shell. This script supports bash and zsh only."
fi
