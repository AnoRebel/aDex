# aDex-UI Theme System Guide

This guide covers creating, customizing, and importing themes for aDex-UI.

## Table of Contents

- [Theme System Overview](#theme-system-overview)
- [Built-in Themes](#built-in-themes)
- [Creating Custom Themes](#creating-custom-themes)
- [Theme Structure](#theme-structure)
- [CSS Variables Reference](#css-variables-reference)
- [Importing Legacy Themes](#importing-legacy-themes)
- [Theme Development](#theme-development)
- [Best Practices](#best-practices)
- [Troubleshooting](#theme-troubleshooting)

## Theme System Overview

aDex-UI uses a modern CSS variable-based theming system that provides:

- **Live theme switching** without application restart
- **CSS variable foundation** for consistent styling
- **Legacy theme compatibility** with eDEX-UI theme import
- **Custom theme support** with easy creation
- **Component-aware styling** with scoped variables

## Built-in Themes

### 1. Default Theme
- **Description**: Clean, modern theme with blue accents
- **Best for**: General use, development work
- **Colors**: Blue primary, neutral grays, white backgrounds

### 2. Cyberpunk Theme
- **Description**: Futuristic theme with neon colors and dark backgrounds
- **Best for**: Sci-fi enthusiasts, dark environments
- **Colors**: Pink/cyan neon, dark backgrounds, high contrast

### 3. Retro Theme
- **Description**: 8-bit retro computer aesthetic with green phosphor
- **Best for**: Terminal purists, vintage computing feel
- **Colors**: Green phosphor on black, amber accents

### 4. Minimal Theme
- **Description**: Clean, minimal theme with subtle colors
- **Best for**: Focus-oriented work, reduced distraction
- **Colors**: Grayscale, subtle blue accents, muted colors

## Creating Custom Themes

### Quick Start

1. **Create theme directory**:
   ```bash
   mkdir ~/.config/aDex-UI/themes/my-theme
   cd ~/.config/aDex-UI/themes/my-theme
   ```

2. **Create theme manifest** (`theme.json`):
   ```json
   {
     "name": "My Theme",
     "displayName": "My Custom Theme",
     "description": "A custom theme for aDex-UI",
     "version": "1.0.0",
     "author": "Your Name",
     "colors": {
       "primary": "#1976d2",
       "secondary": "#424242",
       "background": "#121212",
       "surface": "#1e1e1e",
       "text": "#ffffff"
     }
   }
   ```

3. **Create CSS file** (`theme.css`):
   ```css
   :root {
     --primary-color: #1976d2;
     --primary-hover: #1565c0;
     --primary-active: #0d47a1;

     --secondary-color: #424242;
     --secondary-hover: #616161;
     --secondary-active: #212121;

     --background-color: #121212;
     --surface-color: #1e1e1e;
     --card-color: #2a2a2a;

     --text-primary: #ffffff;
     --text-secondary: #b0b0b0;
     --text-muted: #757575;

     --border-color: #333333;
     --border-hover: #444444;

     --success-color: #4caf50;
     --warning-color: #ff9800;
     --error-color: #f44336;
     --info-color: #2196f3;
   }
   ```

4. **Load theme in aDex-UI**:
   - Open Theme Settings
   - Click "Import Theme"
   - Select your theme directory
   - Apply theme

## Theme Structure

### Theme Directory Layout
```
my-theme/
├── theme.json           # Theme manifest
├── theme.css            # CSS variables (required)
├── preview.png          # Theme preview image (optional)
├── sounds/              # Custom audio files (optional)
│   ├── click.wav
│   └── notification.wav
└── fonts/               # Custom fonts (optional)
    ├── terminal.ttf
    └── ui.woff2
```

### Theme Manifest (theme.json)
```json
{
  "name": "my-theme",
  "displayName": "My Custom Theme",
  "description": "A beautiful custom theme",
  "version": "1.0.0",
  "author": "Your Name",
  "homepage": "https://github.com/user/my-theme",
  "license": "MIT",
  "preview": "preview.png",
  "colors": {
    "primary": "#1976d2",
    "secondary": "#424242",
    "accent": "#ff4081"
  },
  "fonts": {
    "terminal": "JetBrains Mono",
    "ui": "Inter"
  },
  "sounds": {
    "enabled": true,
    "pack": "default"
  },
  "compatibility": {
    "minVersion": "1.0.0",
    "maxVersion": null
  }
}
```

### CSS Variables Reference

#### Color Variables
```css
:root {
  /* Primary Colors */
  --primary-color: #1976d2;        /* Main brand color */
  --primary-hover: #1565c0;        /* Hover state */
  --primary-active: #0d47a1;       /* Active/pressed state */
  --primary-light: #e3f2fd;         /* Light variant */
  --primary-dark: #0d47a1;          /* Dark variant */

  /* Secondary Colors */
  --secondary-color: #424242;      /* Secondary brand color */
  --secondary-hover: #616161;      /* Hover state */
  --secondary-active: #212121;     /* Active/pressed state */

  /* Background Colors */
  --background-color: #121212;     /* Main background */
  --surface-color: #1e1e1e;         /* Card/surface background */
  --card-color: #2a2a2a;            /* Card background */
  --overlay-color: rgba(0,0,0,0.8);  /* Modal/overlay background */

  /* Text Colors */
  --text-primary: #ffffff;          /* Primary text */
  --text-secondary: #b0b0b0;        /* Secondary text */
  --text-muted: #757575;           /* Muted/disabled text */
  --text-link: #90caf9;             /* Link text */

  /* Border Colors */
  --border-color: #333333;          /* Standard border */
  --border-hover: #444444;          /* Hover border */
  --border-focus: #1976d2;          /* Focus border */

  /* Status Colors */
  --success-color: #4caf50;         /* Success/green */
  --warning-color: #ff9800;         /* Warning/orange */
  --error-color: #f44336;           /* Error/red */
  --info-color: #2196f3;            /* Info/blue */

  /* Terminal Specific */
  --terminal-bg: #000000;           /* Terminal background */
  --terminal-fg: #00ff00;           /* Terminal foreground */
  --terminal-cursor: #00ff00;       /* Cursor color */
  --terminal-selection: #003300;    /* Selection background */
}
```

#### Typography Variables
```css
:root {
  /* Font Families */
  --font-family-mono: 'JetBrains Mono', 'Fira Code', monospace;
  --font-family-sans: 'Inter', 'Roboto', sans-serif;
  --font-family-display: 'Orbitron', monospace;

  /* Font Sizes */
  --font-size-xs: 0.75rem;
  --font-size-sm: 0.875rem;
  --font-size-base: 1rem;
  --font-size-lg: 1.125rem;
  --font-size-xl: 1.25rem;
  --font-size-2xl: 1.5rem;
  --font-size-3xl: 1.875rem;

  /* Font Weights */
  --font-weight-light: 300;
  --font-weight-normal: 400;
  --font-weight-medium: 500;
  --font-weight-semibold: 600;
  --font-weight-bold: 700;

  /* Line Heights */
  --line-height-tight: 1.25;
  --line-height-normal: 1.5;
  --line-height-relaxed: 1.75;
}
```

#### Spacing Variables
```css
:root {
  --spacing-1: 0.25rem;   /* 4px */
  --spacing-2: 0.5rem;    /* 8px */
  --spacing-3: 0.75rem;   /* 12px */
  --spacing-4: 1rem;      /* 16px */
  --spacing-5: 1.25rem;   /* 20px */
  --spacing-6: 1.5rem;    /* 24px */
  --spacing-8: 2rem;      /* 32px */
  --spacing-10: 2.5rem;   /* 40px */
  --spacing-12: 3rem;     /* 48px */
  --spacing-16: 4rem;     /* 64px */
}
```

#### Effects Variables
```css
:root {
  /* Shadows */
  --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  --shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
  --shadow-xl: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);

  /* Border Radius */
  --radius-none: 0;
  --radius-sm: 0.125rem;
  --radius: 0.25rem;
  --radius-md: 0.375rem;
  --radius-lg: 0.5rem;
  --radius-xl: 0.75rem;
  --radius-2xl: 1rem;
  --radius-full: 9999px;

  /* Transitions */
  --transition-fast: 150ms ease-in-out;
  --transition-normal: 250ms ease-in-out;
  --transition-slow: 350ms ease-in-out;

  /* Z-Index Scale */
  --z-dropdown: 1000;
  --z-sticky: 1020;
  --z-fixed: 1030;
  --z-modal-backdrop: 1040;
  --z-modal: 1050;
  --z-popover: 1060;
  --z-tooltip: 1070;
}
```

## Importing Legacy Themes

### eDEX-UI Theme Import

1. **Locate legacy theme files**:
   ```bash
   # eDEX-UI themes are typically in:
   ~/.config/eDEX-ui/themes/
   ```

2. **Use the theme converter**:
   ```bash
   # Using aDex-UI built-in converter
   task theme:import --legacy ~/.config/eDEX-ui/themes/theme.css
   ```

3. **Manual conversion** (theme.css example):
   ```css
   /* Legacy eDEX-UI theme */
   :root {
     --bg-color: #0f0f23;
     --bg-grid: #2a2a3e;
     --main-fg: #f1f1f1;
     --secondary-fg: #b0b0b0;
     --cyan: #00bcd4;
     --magenta: #ff4081;
     --yellow: #ffeb3b;
   }
   ```

4. **Converted theme structure**:
   ```json
   {
     "name": "legacy-edex",
     "displayName": "Legacy eDEX-UI",
     "converted": true,
     "originalVersion": "2.2.4",
     "colors": {
       "background": "#0f0f23",
       "surface": "#2a2a3e",
       "text": "#f1f1f1",
       "textSecondary": "#b0b0b0",
       "cyan": "#00bcd4",
       "magenta": "#ff4081",
       "yellow": "#ffeb3b"
     }
   }
   ```

### Theme Compatibility

The theme converter handles:

- **Color mapping**: Legacy colors to modern CSS variables
- **Font adjustments**: Terminal font configurations
- **Layout compatibility**: Spacing and sizing adjustments
- **Feature mapping**: Legacy features to modern equivalents

## Theme Development

### Development Setup

1. **Enable theme development mode**:
   ```bash
   export ADEX_THEME_DEV=1
   task dev
   ```

2. **Create development theme**:
   ```bash
   task theme:create --name dev-theme --template default
   ```

3. **Live reloading**:
   - Edit theme CSS files
   - Changes apply automatically in development mode
   - Use browser dev tools for inspection

### Testing Themes

1. **Theme validation**:
   ```bash
   task theme:validate --path ./my-theme
   ```

2. **Theme preview**:
   ```bash
   task theme:preview --name my-theme
   ```

3. **Compatibility testing**:
   ```bash
   task theme:test --all-platforms
   ```

### Theme Components

#### Terminal Styling
```css
.terminal-panel {
  background-color: var(--terminal-bg);
  color: var(--terminal-fg);
  border: 1px solid var(--border-color);
  border-radius: var(--radius);
}

.terminal-cursor {
  background-color: var(--terminal-cursor);
  animation: blink 1s infinite;
}

.terminal-selection {
  background-color: var(--terminal-selection);
}
```

#### System Monitor Styling
```css
.system-monitor {
  background-color: var(--surface-color);
  border: 1px solid var(--border-color);
}

.chart-canvas {
  background-color: transparent;
}

.metric-value {
  color: var(--text-primary);
  font-family: var(--font-family-mono);
}
```

#### File Browser Styling
```css
.file-browser {
  background-color: var(--surface-color);
}

.file-item {
  color: var(--text-secondary);
  transition: color var(--transition-fast);
}

.file-item:hover {
  color: var(--text-primary);
  background-color: var(--background-color);
}

.file-icon {
  color: var(--primary-color);
}
```

## Best Practices

### Color Design

1. **Contrast ratios**: Ensure WCAG AA compliance (4.5:1 minimum)
2. **Consistent hierarchy**: Use clear color relationships
3. **Accessibility**: Test with colorblind simulators
4. **Dark mode**: Most users prefer dark themes for terminal work

### Performance

1. **CSS efficiency**: Use CSS variables for consistency
2. **Minimal overrides**: Only override necessary variables
3. **Image optimization**: Use efficient preview images
4. **File size**: Keep theme files under 50KB

### User Experience

1. **Intuitive naming**: Use descriptive color names
2. **Smooth transitions**: Use consistent transition timing
3. **Visual feedback**: Clear hover and active states
4. **Readability**: Prioritize legibility over aesthetics

### Maintenance

1. **Version control**: Use semantic versioning
2. **Documentation**: Include change logs
3. **Testing**: Test across all supported platforms
4. **Updates**: Keep themes compatible with new versions

## Theme Troubleshooting

### Common Issues

#### Theme Not Applying
**Problem**: Theme changes don't appear after switching
**Solution**:
1. Check theme syntax and CSS validity
2. Verify theme manifest format
3. Restart application if necessary
4. Clear theme cache: `rm -rf ~/.config/aDex-UI/cache`

#### Colors Not Updating
**Problem**: Some components retain old colors
**Solution**:
1. Check if component uses inline styles
2. Verify CSS variable names
3. Check component specificity
4. Use browser dev tools to inspect styles

#### Preview Image Not Showing
**Problem**: Theme preview missing in theme selector
**Solution**:
1. Ensure preview.png exists in theme directory
2. Check image dimensions (recommended 800x600)
3. Verify image format (PNG or JPEG)
4. Check file permissions

#### Legacy Theme Import Fails
**Problem**: eDEX-UI theme import fails
**Solution**:
1. Check theme file format and syntax
2. Verify theme file permissions
3. Use manual conversion if automatic fails
4. Check for missing color definitions

### Debug Mode

Enable theme debugging:
```bash
export ADEX_THEME_DEBUG=1
./aDex-UI
```

Debug output shows:
- Theme loading process
- CSS variable resolution
- Theme validation results
- Performance metrics

### CSS Variable Inspection

Use browser dev tools to inspect:
1. Computed styles for variables
2. Variable inheritance and overrides
3. Specificity conflicts
4. Performance impact

### Theme Validation

Validate themes programmatically:
```bash
task theme:check --path ./my-theme --strict
```

Validation includes:
- CSS syntax checking
- Required variable presence
- Color contrast validation
- File structure verification

---

## Additional Resources

- **Color Palettes**: [Coolors](https://coolors.co/), [Adobe Color](https://color.adobe.com/)
- **CSS Tools**: [CSS Variables Generator](https://cssvars.com/)
- **Accessibility**: [WebAIM Contrast Checker](https://webaim.org/resources/contrastchecker/)
- **Inspiration**: [Dribbble](https://dribbble.com/), [Behance](https://www.behance.net/)

## Community Themes

Share your themes with the community:

1. **GitHub**: Create a repository with your theme
2. **Theme Directory**: Submit to aDex-UI theme registry
3. **Documentation**: Include installation and usage instructions
4. **Support**: Provide maintenance and updates

---

Happy theming! 🎨