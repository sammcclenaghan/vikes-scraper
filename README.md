# Vikes Scraper

A command-line tool to scrape UVic course information from Banner and generate course schedules.

## Usage

### Basic Usage

```bash
# Fetch information about a single course
./vikes-scraper -course CSC 110

# Fetch information about multiple courses
./vikes-scraper -courses CSC 110 MATH 100 PHYS 110

# Fetch all courses and export to CSV
./vikes-scraper -all

# Perform a dry run without saving data
./vikes-scraper -all -dry-run
```

### Specifying a Semester

By default, the scraper uses term code `202501` (Spring 2025). You can specify a different semester using the `-semester` flag:

```bash
# Fetch course information for Fall 2025
./vikes-scraper -course -semester=202509 CSC 110

# Fetch all courses for Summer 2025
./vikes-scraper -all -semester=202505
```

### Schedule Generator

Generate possible non-conflicting schedules for multiple courses:

```bash
# Generate schedules for multiple courses
./vikes-scraper -schedule -courses CSC 110 MATH 100 PHYS 110

# Generate schedules for different semester
./vikes-scraper -schedule -semester=202509 -courses CSC 110 MATH 100 STAT 260

# Limit the number of generated schedules (default is 5)
./vikes-scraper -schedule -max-schedules=10 -courses CSC 110 MATH 100 ENGR 130
```

The schedule generator will:
- Find all possible combinations of lecture/lab/tutorial sections
- Eliminate schedules with time conflicts
- Score schedules based on desirable features (fewer early mornings, days off, etc.)
- Display the top-ranked schedules

## Term Codes

UVic uses a 6-digit term code system:
- First 4 digits: year (e.g., 2025)
- Last 2 digits: term code
  - `01` = Spring (January)
  - `05` = Summer (May)
  - `09` = Fall (September)

## Output

The program outputs course information including:
- Course reference number (CRN)
- Subject and course number
- Section
- Title
- Professor and contact information
- Schedule and location
- Enrollment information
- Credit hours

## Building

```bash
go build -o vikes-scraper
```