# Nuxt Image Optimization for File System

This document describes the Nuxt Image optimizations implemented for the file system components in Dex-UI.

## Overview

The file system components now use Nuxt Image for optimal performance, including automatic format detection, lazy loading, CDN integration, and responsive image handling.

## Features

### 🚀 Performance Optimizations

- **Automatic Format Detection**: Automatically detects browser support for AVIF, WebP, and falls back to JPEG/PNG
- **Size-Based Optimization**: Different quality and blur settings based on icon size
- **Lazy Loading**: Images are loaded lazily when they enter the viewport
- **Placeholder Images**: Blur placeholders for smooth loading experience
- **CDN Integration**: Support for custom CDN URLs for faster delivery

### 📱 Responsive Design

- **Multi-Density Support**: Loads appropriate images for different pixel densities (1x, 1.5x, 2x)
- **Adaptive Quality**: Adjusts quality based on network conditions
- **Memory Awareness**: Reduces quality on low-memory devices

### 📊 Performance Monitoring

- **Load Time Tracking**: Monitors image load times and averages
- **Error Tracking**: Records and handles image loading errors
- **Success Rate**: Tracks overall image loading success rates

## Usage

### Basic File Icon

```vue
<template>
  <FileIcon
    :file="fileEntry"
    :size="48"
    :show-preview="true"
    theme="auto"
  />
</template>
```

### Advanced Configuration

```vue
<template>
  <FileIcon
    :file="fileEntry"
    :size="64"
    :show-preview="true"
    :override-image-config="{
      format: 'webp',
      quality: 90,
      loading: 'eager'
    }"
  />
</template>
```

### Global Configuration

```ts
// In your app setup
import { nuxtImageConfig } from '~/composables/useNuxtImageConfig'

// Update global settings
nuxtImageConfig.updateConfig({
  format: 'auto',
  quality: 85,
  enableWebP: true,
  enableAVIF: true,
  cdnUrl: 'https://cdn.example.com'
})
```

## Configuration Options

### Global Configuration

```typescript
interface ImageConfig {
  format: 'auto' | 'avif' | 'webp' | 'jpg' | 'png'
  quality: number           // 1-100
  loading: 'lazy' | 'eager'
  placeholder: boolean
  cdnUrl: string
  enableWebP: boolean
  enableAVIF: boolean
  optimizeForSize: boolean
}
```

### Size-Based Optimization

| Size | Quality | Blur | Use Case |
|------|--------|------|----------|
| 16px | 60% | 2px | Small icons |
| 24px | 70% | 3px | Standard icons |
| 48px | 80% | 4px | Medium previews |
| 64px+ | 85% | 6px | Large previews |

### File Type Quality Adjustment

- **PNG/SVG**: +10 quality (graphics benefit from higher quality)
- **WebP**: +5 quality (more efficient format)
- **AVIF**: +10 quality (most efficient format)
- **JPEG**: Standard quality (photos)

## Adaptive Behavior

### Network-Based Adjustments

The system automatically adjusts image settings based on:

- **Data Saver Mode**: Reduces quality to 60%, enables lazy loading
- **Slow Networks** (2G/3G): Reduces quality, disables AVIF
- **Fast Networks** (4G+): Uses optimal settings

### Device-Based Adjustments

- **Low Memory** (<4GB): Reduces quality to 65%
- **Medium Memory** (4-8GB): Quality 75%, lazy loading
- **High Memory** (>8GB): Optimal settings

### User Preference Support

- **Reduced Motion**: Disables animations that affect performance
- **Page Visibility**: Pauses operations when tab is hidden

## Components

### FileIcon Component

The main component that uses Nuxt Image for file thumbnails.

**Props:**
- `file`: FileSystemEntry - The file to display
- `size`: number - Icon size in pixels (default: 24)
- `showPreview`: boolean - Enable thumbnail preview (default: true)
- `showStatus`: boolean - Show loading/error indicators (default: true)
- `showExtension`: boolean - Show file extension badge (default: false)
- `color`: string - Custom icon color
- `theme`: 'light' | 'dark' | 'auto' - Theme preference (default: 'auto')
- `style`: 'emoji' | 'svg' | 'mixed' - Icon style preference (default: 'mixed')
- `overrideImageConfig`: Partial configuration override

### Composables

#### `useNuxtImageConfig`

Global configuration manager for Nuxt Image settings.

```typescript
const {
  config,              // Current configuration
  updateConfig,         // Update configuration
  resetConfig,          // Reset to defaults
  getOptimalFormat,     // Get best format for browser
  getSizeConfig,        // Get size-specific config
  getOptimalQuality,    // Get quality for file type
  getPlaceholderConfig, // Get placeholder settings
  shouldUseLazyLoading, // Determine loading strategy
  getCdnUrl,           // Generate CDN URLs
  getResponsiveSizes,  // Generate responsive sizes
  getImageAttributes,  // Get complete image attributes
  performanceStats,    // Performance metrics
  performanceReport,   // Performance report
  recordImageLoad,    // Record successful load
  recordImageError,    // Record load error
  isWebPSupported,     // WebP support status
  isAVIFSupported      // AVIF support status
} = useNuxtImageConfig()
```

## Performance Metrics

### Tracking

The system tracks:
- **Images Loaded**: Total count of successfully loaded images
- **Images Failed**: Total count of failed loads
- **Average Load Time**: Rolling average of image load times
- **Error Rate**: Percentage of failed loads
- **Success Rate**: Percentage of successful loads

### Monitoring

Access performance metrics:

```ts
import { nuxtImageConfig } from '~/composables/useNuxtImageConfig'

console.log('Performance Report:', nuxtImageConfig.performanceReport.value)
```

### Core Web Vitals

The plugin monitors Core Web Vitals that affect image loading:
- **LCP (Largest Contentful Paint)**: When the largest image loads
- **CLS (Cumulative Layout Shift)**: Layout shifts from image loading
- **FID (First Input Delay)**: Interactivity timing

## CDN Integration

### Setup

Configure CDN in your `.env` file:

```env
NUXT_PUBLIC_CDN_URL=https://cdn.example.com
# or
NUXT_PUBLIC_IMAGE_CDN_URL=https://images.example.com
```

### Usage

The system automatically prepends the CDN URL to image paths:

```typescript
// Original: /api/filesystem/thumbnail?path=/home/user/image.jpg
// With CDN: https://cdn.example.com/api/filesystem/thumbnail?path=/home/user/image.jpg
```

## Best Practices

### 1. Use Appropriate Sizes

```vue
<!-- Small icons -->
<FileIcon :size="16" />

<!-- Standard icons -->
<FileIcon :size="24" />

<!-- Large previews -->
<FileIcon :size="48" />
```

### 2. Configure Loading Strategy

```vue
<!-- Above the fold -->
<FileIcon :override-image-config="{ loading: 'eager' }" />

<!-- Below the fold -->
<FileIcon :override-image-config="{ loading: 'lazy' }" />
```

### 3. Adjust Quality for Use Case

```vue
<!-- High quality for important images -->
<FileIcon :override-image-config="{ quality: 95 }" />

<!-- Lower quality for decorative images -->
<FileIcon :override-image-config="{ quality: 60 }" />
```

### 4. Monitor Performance

```ts
// Check performance metrics periodically
onMounted(() => {
  setInterval(() => {
    const report = nuxtImageConfig.performanceReport.value
    if (report.errorRate > 5) {
      console.warn('High image error rate detected:', report.errorRate)
    }
  }, 30000) // Check every 30 seconds
})
```

## Troubleshooting

### Images Not Loading

1. **Check Console**: Look for image loading errors
2. **Verify Path**: Ensure thumbnail URLs are correct
3. **Check CDN**: Verify CDN configuration if used
4. **Test Format**: Try disabling AVIF/WebP for compatibility

### Performance Issues

1. **Reduce Quality**: Lower image quality for faster loading
2. **Enable Lazy Loading**: Only load visible images
3. **Check Network**: Monitor for slow connections
4. **Review Settings**: Adjust global configuration

### Format Support Issues

1. **Browser Compatibility**: Some browsers don't support AVIF
2. **Server Configuration**: Ensure server supports modern formats
3. **Fallback**: System automatically falls back to JPEG/PNG

## Future Enhancements

### Planned Features

- **WebP Animation Support**: Animated WebP thumbnails
- **Progressive JPEGs**: Progressive image loading
- **Image Preloading**: Intelligent preloading of critical images
- **Custom Filters**: Apply filters and effects client-side
- **Image Caching**: Enhanced client-side caching strategies

### Performance Improvements

- **Service Worker Integration**: Cache images for offline use
- **Image Compression**: Client-side compression options
- **Batch Loading**: Optimize loading of multiple images
- **Resource Prioritization**: Prioritize important images