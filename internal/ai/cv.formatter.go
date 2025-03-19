package ai

import "fmt"

func FormatForATS(cvContent, jobDescription, apiKey string) (string, error) {
	if cvContent == "" {
		return "", fmt.Errorf("CV content is empty")
	}
	if jobDescription == "" {
			return "", fmt.Errorf("job description is empty")
	}

	prompt := fmt.Sprintf(`You are an expert CV/resume formatter that specializes in creating ATS-friendly resumes.
	Your task is to reformat the provided CV to optimize it for Applicant Tracking Systems (ATS) based on the job description.
	Follow these guidelines:
	1. Use a clean, standard formatting
	2. Include relevant keywords from the job description
	3. Quantify achievements where possible
	4. Emphasize skills and experiences that match the job requirements
	5. Format the sections in a standardized way: Contact Information, Professional Summary, Work Experience, Skills, Education
	6. Return the updated CV text that is optimized for the job description.
	Here is the job description:
	%s
	Here is the CV to optimize:
	%s
	Please provide only the optimized CV content in a clean format that would work well for ATS systems.
	`, jobDescription, cvContent)
	
	return callGemini(prompt, apiKey)
}

func RoastCV(cvContent, apiKey string) (string, error) {
	if cvContent == "" {
		return "", fmt.Errorf("CV content is empty")
	}

	prompt := fmt.Sprintf(`You are a brutally honest CV reviewer. Your job is to "roast" the following CV by:
	1. Identifying weak, generic, or cliché language
	2. Pointing out missing or vague quantifiable achievements
	3. Highlighting formatting or structure issues
	4. Noting overused buzzwords or jargon
	5. Suggesting specific improvements
	Be direct, somewhat humorous, but ultimately constructive. The goal is to help the person improve their CV through honest feedback.
	Here is the CV to roast:
	%s
	Provide your feedback as bullet points with clear, actionable suggestions for improvement.
	`, cvContent)

	return callGemini(prompt, apiKey)
}
