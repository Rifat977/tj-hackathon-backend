# Assets and Product Management

This document describes the assets management and product seeding functionality in the Go Fiber Boilerplate.

## Assets Structure

The project includes the following assets:

```
assets/
├── products.json          # Product data (17,000+ products)
├── images/
│   ├── Products/          # Product images (1.jpg to 20.jpg)
│   └── Banners/           # Banner images (banner1.jpg, banner2.jpg, banner3.jpg)
```

## Product Model

The Product model has been updated to include all fields from the JSON data:

- `Index`: Product index number
- `Name`: Product name
- `Description`: Full HTML description
- `ShortDescription`: Brief description
- `Brand`: Product brand
- `Category`: Product category (string)
- `Price`: Product price
- `Currency`: Price currency (default: USD)
- `Stock`: Available stock quantity
- `EAN`: European Article Number
- `Color`: Product color
- `Size`: Product size
- `Availability`: Stock availability status
- `Image`: Image filename
- `InternalID`: Unique internal identifier
- `Slug`: URL-friendly slug
- `SKU`: Stock keeping unit
- `CategoryModel`: Related category model (optional)

## API Endpoints

### Product Seeding

#### Seed Products from JSON
```http
POST /api/seed/products
```
Seeds the database with products from `assets/products.json`. This will process all 17,000+ products and create appropriate categories automatically.

#### Clear Products Data
```http
DELETE /api/seed/clear
```
Clears all products and categories from the database. This action cannot be undone.

### Product Management

#### Get Products
```http
GET /api/products?page=1&limit=10&category_id=1
```

#### Get Product by ID
```http
GET /api/products/{id}
```

#### Search Products
```http
GET /api/products/search?q=search_term&category=category_slug&min_price=10&max_price=100
```

#### Get Categories
```http
GET /api/categories
```

#### Get Products by Category
```http
GET /api/categories/{id}/products
```

### Banner Management

#### Get Banners
```http
GET /api/banners
```
Returns available banner images with their URLs.

## Static File Serving

Product and banner images are served statically:

- Product images: `/assets/images/Products/{filename}`
- Banner images: `/assets/images/Banners/{filename}`

## Usage Examples

### 1. Seed Products from JSON
```bash
curl -X POST http://localhost:3000/api/seed/products
```

### 2. Clear Products Data
```bash
curl -X DELETE http://localhost:3000/api/seed/clear
```

### 3. Get Products with Images
```bash
curl http://localhost:3000/api/products
```

### 4. Get Banners
```bash
curl http://localhost:3000/api/banners
```

## Notes

- Product images are automatically served from the `/assets/images/Products/` directory
- Banner images are served from `/assets/images/Banners/` directory
- The seeding process is idempotent - running it multiple times won't create duplicates
- Categories are automatically created based on the product categories in the JSON
- All 17,000+ products from the JSON file are processed and stored in the database
- The clear function properly removes all products and categories while respecting foreign key constraints 