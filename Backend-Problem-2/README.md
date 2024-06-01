# Top Products HTTP Microservice

This project is a microservice that provides an API to fetch the top N products from multiple e-commerce companies within a specified category and price range. Users can also sort the results based on various criteria.

## Features

- Fetch top N products from five e-commerce companies
- Pagination support for large result sets
- Sorting by price, rating, discount, and company
- Retrieve detailed information about a specific product

## Endpoints

### 1. Get Top N Products

- **URL**: `/categories/{categoryname}/products`
- **Method**: `GET`
- **Query Parameters**:
  - `top`: Number of top products to retrieve (default is 10)
  - `minPrice`: Minimum price range
  - `maxPrice`: Maximum price range
  - `sortBy`: Field to sort by (`price`, `rating`, `discount`, `company`)
  - `order`: Sort order (`asc` for ascending, `desc` for descending)
  - `page`: Page number for pagination

#### Example

```sh
GET /categories/Laptop/products?top=10&minPrice=100&maxPrice=2000&sortBy=price&order=asc&page=1
