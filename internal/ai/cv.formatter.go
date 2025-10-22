package ai

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/Emmanuella-codes/burnished-microservice/internal/dtos"
)

func ParseAndOptimizeCV(cvContent, jobDescription, apiKey string) (*dtos.Resume, error) {
	if cvContent == "" {
		return nil, fmt.Errorf("CV content is empty")
	}
	if jobDescription == "" {
			return nil, fmt.Errorf("job description is empty")
	}

	prompt := fmt.Sprintf(`You are an expert resume parser and ATS optimizer. Parse the resume AND optimize it for the job description in ONE step.
	
	SECTION DETECTION - Recognize these variations:
	- Experience: Work History, Employment, Professional Experience, Career History
	- Education: Academic Background, Qualifications, Academic History
	- Skills: Technical Skills, Core Competencies, Expertise, Proficiencies
	- Projects: Portfolio, Personal Projects, Side Projects
	- Awards: Honors, Achievements, Recognition, Certifications
	- Summary: Profile, About, Professional Summary, Objective

	PARSING RULES:
	1. Normalize dates to "MMM YYYY" format (use "Present" for current roles)
	2. Extract URLs separately from display names
	3. Extract all bullet points as array items
	4. Initialize empty arrays for missing sections
	5. Infer skills from entire document if no dedicated section

	OPTIMIZATION RULES:
	1. Match keywords from job description naturally (don't stuff)
	2. Quantify achievements with metrics (%%, $, X, numbers)
	3. Use strong action verbs: Led, Developed, Implemented, Achieved, Increased, Reduced
	4. Tailor profile summary to match the target role
	5. Reorder/emphasize relevant experiences and skills
	6. Add missing but relevant skills from job description IF candidate has related experience
	7. Enhance bullet points with impact and results
	8. NEVER fabricate experience - only enhance existing content
	9. Keep all information truthful and grounded in original resume

	Here is the job description:
	%s

	Here is the Resume to parse optimize:
	%s

	OUTPUT (JSON only, no markdown):
	{
		"header": {
			"fullname": "", "jobTitle": "", "location": "", "email": "",
			"phone": "", "linkedin": "", "linkedinUrl": "", "github": "",
			"githubUrl": "", "website": "", "websiteUrl": ""
		},
		"profileSummary": "Optimized 2-3 sentence summary tailored to job",
		"skills": [{"title": "Category", "values": ["relevant skills first"]}],
		"experiences": [{
			"company": "", "occupation": "", "startDate": "MMM YYYY",
			"endDate": "MMM YYYY", "location": "",
			"desc": ["Enhanced bullets with metrics and impact"]
		}],
		"education": [{
			"degree": "", "institution": "", "startDate": "MMM YYYY",
			"endDate": "MMM YYYY", "location": "", "desc": []
		}],
		"projects": [{"title": "", "link": "", "subtitle": "", "desc": ["Enhanced descriptions"]}],
		"awards": [{"title": "", "link": "", "subtitle": "", "date": "MMM YYYY", "desc": []}],
		"sectionOrder": ["header", "profileSummary", "experiences", "education", "skills", "projects", "awards"]
	}

	Return ONLY the optimized JSON:`, jobDescription, cvContent)

	response, err := callGemini(prompt, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to call AI: %w", err)
	}

	cleanedResponse := cleanMarkdownJSON(response)

	var resume dtos.Resume
	if err := json.Unmarshal([]byte(cleanedResponse), &resume); err != nil {
		log.Printf("Failed to unmarshal JSON. Cleaned response: %s", cleanedResponse)
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	ValidateAndFillMissingSections(&resume)
	return &resume, nil
}

func RoastCV(cvContent, apiKey string) (string, error) {
	if cvContent == "" {
		return "", fmt.Errorf("CV content is empty")
	}

	prompt := fmt.Sprintf(`You are a SAVAGE CV reviewer who has seen thousands of terrible resumes. Your job is to absolutely DEMOLISH this CV with brutal honesty. DO NOT hold back.
		IGNORE these when roasting:
	- Dates (employment dates, education dates, etc.)
	- Contact details in the header (GitHub, LinkedIn, website links)
	- Name and basic contact info

	ROAST MERCILESSLY:
	1. DESTROY weak, generic language - "responsible for", "worked on", "helped with" deserve no mercy
	2. OBLITERATE vague achievements - if there are no numbers, percentages, or concrete results, tear it apart
	3. SAVAGE the boring bullet points - if it sounds like a job description copy-paste, roast it hard
	4. DEMOLISH buzzword spam - "synergy", "rockstar", "guru", "passionate" etc. Show no mercy
	5. ANNIHILATE poor formatting - walls of text, inconsistent styling, amateur mistakes
	6. BRUTALIZE the lack of impact - if they're just listing tasks instead of achievements, destroy them
	7. TORCH generic summaries - "hard-working team player seeking opportunities" deserves mockery
	8. FLAME missing specifics - technologies without context, projects without outcomes

	Your tone should be:
	- Sarcastic and cutting
	- Brutally honest, borderline offensive (but not discriminatory)
	- Use humor, but make it HURT
	- Compare bad examples to what they SHOULD say
	- Don't sugarcoat ANYTHING

	Here is the CV to absolutely roast:
	%s

	Format your roast as:
	1. Start with a savage one-liner summary of the overall CV
	2. Section-by-section destruction with specific examples of what sucks
	3. A "wake-up call" conclusion that's motivational but brutal

	Make them FEEL the pain of their mediocre CV. No participation trophies here.`, cvContent)

	return callGemini(prompt, apiKey)
}

func ValidateAndFillMissingSections(resume *dtos.Resume) {
	if len(resume.SectionOrder) == 0 {
		resume.SectionOrder = []string{
			"header", "profileSummary", "experiences",
			"education", "skills", "projects", "awards",
		}
	}

	if resume.Skills == nil {
		resume.Skills = []dtos.Skills{}
	}
	if resume.Experiences == nil {
		resume.Experiences = []dtos.Experience{}
	}
	if resume.Education == nil {
		resume.Education = []dtos.Education{}
	}
	if resume.Projects == nil {
		resume.Projects = []dtos.Project{}
	}
	if resume.Awards == nil {
		resume.Awards = []dtos.Award{}
	}
}

func cleanMarkdownJSON(response string) string {
	// Remove leading/trailing whitespace
	response = strings.TrimSpace(response)
	
	// Remove ```json and ``` markers
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	
	// Remove any remaining whitespace
	return strings.TrimSpace(response)
}
