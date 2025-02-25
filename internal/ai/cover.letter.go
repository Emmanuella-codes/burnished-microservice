package ai

import "fmt"

func GenerateCoverLetter(cvData []byte, jobDescription, apiKey string) (string, error) {
	cvText := string(cvData)
	prompt := fmt.Sprintf(`Based on the following resume/CV and job description, please create a compelling cover letter.
	The cover letter should:
	1. Be personalized based on the candidate's experience in the CV
	2. Address key requirements from the job description
	3. Highlight the most relevant skills and experiences
	4. Show enthusiasm for the role and company
	5. Be professional but conversational in tone
	6. Be around 300-400 words in length
	Job Description:
	%s
	Candidate's CV:
	%s
	Please write a complete cover letter that the candidate can use or adapt.
	`, jobDescription, cvText)

	return  callGemini(prompt, apiKey)
}