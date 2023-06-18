package prompt

import "github.com/manifoldco/promptui"

func Text(label string) (string, error) {
	prompt := promptui.Prompt{
		Label: label,
	}

	return prompt.Run()
}
